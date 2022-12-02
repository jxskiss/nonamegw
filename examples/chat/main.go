package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/jxskiss/gopkg/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/jxskiss/nonamegw/pkg/zlog"
	"github.com/jxskiss/nonamegw/proto/bizapi"
	"github.com/jxskiss/nonamegw/proto/brokersvc"
	"github.com/jxskiss/nonamegw/proto/protocol"
)

func main() {
	logger, prop, _ := zlog.NewLogger(&zlog.Config{
		Development: true,
		Level:       "debug",
		Format:      "console",
	})
	zlog.ReplaceGlobals(logger, prop)

	chat := NewChat()
	rpcServer := grpc.NewServer()
	rpcImpl := NewRpcImpl(chat)
	bizapi.RegisterBizApiServer(rpcServer, rpcImpl)
	zlog.Infof("starting chat/rpc server listening on %v", cfg.RpcListen)
	go func() {
		ln, err := net.Listen("tcp", cfg.RpcListen)
		if err != nil {
			zlog.Fatalf("failed listen chat/rpc, err= %v", err)
		}
		err = rpcServer.Serve(ln)
		if err != nil {
			zlog.Fatalf("failed serving chat/rpc, err= %v", err)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		zlog.Fatalf("failed get working directory, err= %v", err)
	}
	httpServer := NewHttpImpl(chat, wd+"/web")
	webFiles := httpServer.webFiles()
	router := httprouter.New()
	router.NotFound = webFiles
	router.Handler("GET", "/web/", http.StripPrefix("/web/", webFiles))
	router.Handler("GET", "/token", http.HandlerFunc(httpServer.tokenHandler))
	zlog.Infof("starting chat/http server listening on %v", cfg.HttpListen)
	go func() {
		ln, err := net.Listen("tcp", cfg.HttpListen)
		if err != nil {
			zlog.Fatalf("failed listen chat/http, err= %v", err)
		}
		err = http.Serve(ln, router)
		if err != nil {
			zlog.Fatalf("failed serving chat/http, err= %v", err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	<-exit
	rpcServer.GracefulStop()
}

// ---- Configuration ---- //

var cfg = &Config{
	BrokerAddr:    "127.0.0.1:9432",
	RpcListen:     "127.0.0.1:9433",
	HttpListen:    "127.0.0.1:9434",
	AuthAppId:     1001,
	AuthSecretKey: "dummy_auth_key",
}

type Config struct {
	BrokerAddr    string
	RpcListen     string
	HttpListen    string
	AuthAppId     int64
	AuthSecretKey string
}

func (p *Config) getAuthCredential() *brokersvc.Authorization {
	return &brokersvc.Authorization{
		AppId:     cfg.AuthAppId,
		AccessKey: cfg.AuthSecretKey,
	}
}

var brokerClient brokersvc.BrokerClient

func getBrokerClient() (brokersvc.BrokerClient, error) {
	if brokerClient == nil {
		cc, err := grpc.Dial(cfg.BrokerAddr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		brokerClient = brokersvc.NewBrokerClient(cc)
	}
	return brokerClient, nil
}

// ---- HTTP server ---- //

func NewHttpImpl(chat *Chat, webdir string) *httpImpl {
	return &httpImpl{
		chat:   chat,
		webdir: webdir,
	}
}

type httpImpl struct {
	chat   *Chat
	webdir string
}

func (p *httpImpl) webFiles() http.Handler {
	return http.FileServer(http.Dir(p.webdir))
}

func (p *httpImpl) tokenHandler(resp http.ResponseWriter, req *http.Request) {
	userId := p.chat.NextUserId()
	signReq := &brokersvc.SignTokenRequest{
		Auth:     cfg.getAuthCredential(),
		UserId:   userId,
		DeviceId: 0,
	}
	brokerCli, err := getBrokerClient()
	if err != nil {
		zlog.Errorf("failed get broker client, err= %v", err)
		resp.WriteHeader(500)
		return
	}
	ctx := context.Background()
	signResp, err := brokerCli.SignToken(ctx, signReq)
	if err != nil {
		zlog.Errorf("failed sign token, err= %v", err)
		resp.WriteHeader(500)
		return
	}
	if signResp.Token == "" {
		zlog.Errorf("unexpected empty token from broker service")
		resp.WriteHeader(500)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	json.NewEncoder(resp).Encode(map[string]interface{}{
		"token":     signResp.Token,
		"expire_at": signResp.ExpireAt,
	})
}

// ---- RPC server ---- //

var _ bizapi.BizApiServer = &RpcImpl{}

func NewRpcImpl(chat *Chat) *RpcImpl {
	return &RpcImpl{
		chat: chat,
	}
}

type RpcImpl struct {
	bizapi.UnimplementedBizApiServer

	chat *Chat
}

func (p *RpcImpl) OnMessage(ctx context.Context, request *bizapi.OnMessageRequest) (*bizapi.OnMessageResponse, error) {
	message := request.GetMessage()
	conn := message.GetConn()

	lg := p.chat.lg.With("uid", conn.GetId())
	lg.Info("onMessage: received message")

	// TODO: check this, should we use Packet directly ?
	chatmsg := &Request{}
	err := json.Unmarshal(message.GetContent().GetPayload(), chatmsg)
	if err != nil {
		lg.Infow("onMessage: failed unmarshal message", "error", err)
		return &bizapi.OnMessageResponse{}, nil
	}

	uid := conn.GetId()
	user, ok := p.chat.GetUser(uid)
	if !ok {
		lg.Warn("onMessage: user not found")
		return &bizapi.OnMessageResponse{}, nil
	}

	switch chatmsg.Method {
	case "rename":
		name, ok := chatmsg.Params["name"].(string)
		if !ok {
			if err = user.writeErrorTo(chatmsg, Object{
				"error": "bad params",
			}); err != nil {
				return nil, err
			}
		}
		prev, ok := user.chat.Rename(user, name)
		if !ok {
			if err = user.writeErrorTo(chatmsg, Object{
				"error": "already exists",
			}); err != nil {
				return nil, err
			}
		}
		p.chat.Broadcast("rename", Object{
			"prev": prev,
			"name": name,
			"time": timestamp(),
		})
		if err = user.writeResultTo(chatmsg, nil); err != nil {
			return nil, err
		}
	case "publish":
		chatmsg.Params["author"] = user.name
		chatmsg.Params["time"] = timestamp()

		// hack to demo broadcast using tags
		if room := p.chat.ParseJoinMessage(chatmsg); room != "" {
			user.chat.Join(user, room)
		} else if user.room != "" {
			roomTag := "room:" + user.room
			tags := []string{roomTag}
			chatmsg.Params["room"] = user.room
			p.chat.BroadcastByTag("publish", tags, chatmsg.Params)
		} else {
			p.chat.Broadcast("publish", chatmsg.Params)
		}
	default:
		if err = user.writeErrorTo(chatmsg, Object{
			"error": "not implemented",
		}); err != nil {
			return nil, err
		}
	}
	return &bizapi.OnMessageResponse{}, nil
}

func (p *RpcImpl) OnEvent(ctx context.Context, request *bizapi.OnEventRequest) (*bizapi.OnEventResponse, error) {
	event := request.GetEvent()
	uid := event.GetConn().GetId()
	switch event.Type {
	case protocol.Event_TOUCH:
		p.chat.lg.Infow("onEvent: touching connection", "uid", uid)
	case protocol.Event_CONNECT:
		p.chat.lg.Infow("onEvent: registering connection", "uid", uid)
		p.register(event.GetConn())
	case protocol.Event_DISCONNECT:
		p.chat.lg.Infow("onEvent: removing connection", "uid", uid)
		p.remove(event.GetConn())
	default:
		p.chat.lg.Warnw("onEvent: unsupported event type", "uid", uid, "type", event.GetType())
	}
	return &bizapi.OnEventResponse{}, nil
}

func (p *RpcImpl) register(conn *protocol.Connection) *User {
	uid := conn.GetId()
	user, ok := p.chat.GetUser(uid)
	if ok {
		return user
	}
	return p.chat.Register(conn)
}

func (p *RpcImpl) remove(conn *protocol.Connection) {
	uid := conn.GetId()
	user, ok := p.chat.GetUser(uid)
	if ok {
		p.chat.Remove(user)
	}
}

// ---- chat ---- //

// Object represents generic message parameters.
// In real-world application it is better to avoid such types for better
// performance.
type Object map[string]interface{}

type Request struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params Object `json:"params"`
}

type Response struct {
	ID     int    `json:"id"`
	Result Object `json:"result"`
}

type Error struct {
	ID    int    `json:"id"`
	Error Object `json:"error"`
}

type User struct {
	id   int64
	name string
	chat *Chat
	uid  string
	room string
}

func (u *User) writeErrorTo(req *Request, err Object) error {
	return u.write(Error{
		ID:    req.ID,
		Error: err,
	})
}

func (u *User) writeResultTo(req *Request, result Object) error {
	return u.write(Response{
		ID:     req.ID,
		Result: result,
	})
}

func (u *User) writeNotice(method string, params Object) error {
	return u.write(Request{
		Method: method,
		Params: params,
	})
}

func (u *User) write(x interface{}) error {
	payload, err := json.Marshal(x)
	if err != nil {
		return err
	}
	uids := []string{u.uid}
	req := &brokersvc.PushRequest{
		Auth: cfg.getAuthCredential(),
		Target: &brokersvc.PushTarget{
			Type: brokersvc.PushTarget_CONNECTION,
			Target: &brokersvc.PushTarget_Connections_{
				Connections: &brokersvc.PushTarget_Connections{
					ConnectionIds: uids,
				},
			},
		},
		Content: &protocol.Content{
			Payload: payload,
		},
	}
	u.chat.out <- req
	return nil
}

type Chat struct {
	mu  sync.RWMutex
	seq int64
	us  []*User
	ns  map[string]*User
	ids map[string]*User

	out chan *brokersvc.PushRequest

	lg *zap.SugaredLogger
}

func NewChat() *Chat {
	chat := &Chat{
		ns:  make(map[string]*User),
		ids: make(map[string]*User),
		out: make(chan *brokersvc.PushRequest, 100),
		lg:  zlog.S(),
	}

	for i := 0; i < 3; i++ {
		go chat.writer()
	}
	return chat
}

func (c *Chat) NextUserId() int64 {
	return atomic.AddInt64(&c.seq, 1)
}

// GetUser tells whether a user exists and returns the user.
func (c *Chat) GetUser(uid string) (user *User, ok bool) {
	c.mu.Lock()
	user, ok = c.ids[uid]
	c.mu.RUnlock()
	return
}

func (c *Chat) ParseJoinMessage(req *Request) (room string) {
	text := req.Params["text"]
	parts := strings.SplitN(strings.TrimSpace(text.(string)), " ", 2)
	if len(parts) == 2 && parts[0] == "#join" {
		if _, err := strconv.Atoi(parts[1]); err == nil {
			return parts[1]
		}
	}
	return ""
}

// Join adds user to chat room.
func (c *Chat) Join(user *User, room string) error {
	panic("not implemented")
}

func (c *Chat) BroadcastByTag(method string, tags []string, params Object) error {
	panic("not implemented")
}

func (c *Chat) Register(conn *protocol.Connection) *User {
	c.lg.With("uid", conn.Id, "user_id", conn.UserId).
		Debug("registering chat user")

	uid := conn.Id
	user := &User{
		id:   conn.UserId,
		chat: c,
		uid:  uid,
	}
	c.mu.Lock()
	user.name = c.randName()
	c.us = append(c.us, user)
	c.ns[user.name] = user
	c.ids[user.uid] = user
	c.mu.Unlock()

	user.writeNotice("hello", Object{
		"name": user.name,
	})
	c.Broadcast("greet", Object{
		"name": user.name,
		"time": timestamp(),
	})
	return user
}

// Remove removes user from chat.
func (c *Chat) Remove(user *User) {
	c.mu.Lock()
	removed := c.remove(user)
	c.mu.Unlock()
	if !removed {
		return
	}

	c.Broadcast("goodbye", Object{
		"name": user.name,
		"time": timestamp(),
	})
}

func (c *Chat) Rename(user *User, name string) (prev string, ok bool) {
	c.mu.Lock()
	if _, has := c.ns[name]; !has {
		ok = true
		prev, user.name = user.name, name
		delete(c.ns, prev)
		c.ns[name] = user
	}
	c.mu.Unlock()
	return prev, ok
}

func (c *Chat) Broadcast(method string, params Object) error {
	req := Request{Method: method, Params: params}
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	uids := c.cloneUids()
	pushReq := &brokersvc.PushRequest{
		Auth: nil,
		Target: &brokersvc.PushTarget{
			Type: brokersvc.PushTarget_CONNECTION,
			Target: &brokersvc.PushTarget_Connections_{
				Connections: &brokersvc.PushTarget_Connections{
					ConnectionIds: uids,
				},
			},
			VersionFilters: nil,
		},
		Content: &protocol.Content{
			Payload: payload,
		},
	}
	c.out <- pushReq
	return nil
}

func (c *Chat) cloneUids() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, 0, len(c.us))
	for _, x := range c.us {
		out = append(out, x.uid)
	}
	return out
}

func (c *Chat) writer() {
	for req := range c.out {
		c.lg.Debugw("pushing message", "target", formatTarget(req.GetTarget()))

		brokerCli, err := getBrokerClient()
		if err != nil {
			c.lg.Errorw("failed get broker client", "error", err)
			continue
		}
		ctx, cancel := newTimeoutCtx(3 * time.Second)
		_, err = brokerCli.Push(ctx, req)
		if err != nil {
			c.lg.Errorw("failed call broker.Push", "error", err)
		}
		cancel()
	}
}

func (c *Chat) remove(user *User) bool {
	uid := user.uid
	if _, has := c.ids[uid]; !has {
		return false
	}

	delete(c.ns, user.name)
	delete(c.ids, uid)

	var idx int
	var found = false
	for i, u := range c.us {
		if u.uid == uid {
			idx, found = i, true
			break
		}
	}
	if !found {
		c.lg.DPanic("inconsistent chat state")
	}
	without := make([]*User, len(c.us)-1)
	copy(without[:idx], c.us[:idx])
	copy(without[idx:], c.us[idx+1:])
	c.us = without

	return true
}

func (c *Chat) randName() string {
	var suffix string
	for {
		name := animals[rand.Intn(len(animals))] + suffix
		if _, has := c.ns[name]; !has {
			return name
		}
		suffix += strconv.Itoa(rand.Intn(1000000))
	}
}

func formatTarget(target *brokersvc.PushTarget) string {
	switch target.Type {
	case brokersvc.PushTarget_CONNECTION:
		return fmt.Sprintf("conn_ids: %v", target.GetConnections().ConnectionIds)
	case brokersvc.PushTarget_USER:
		return fmt.Sprintf("user_ids: %v", target.GetUsers().UserIds)
	case brokersvc.PushTarget_USER_DEVICE:
		return fmt.Sprintf("user_devices: %v", target.GetUserDevices().UserDevices)
	case brokersvc.PushTarget_UNAUTHENTICATED_DEVICE:
		return fmt.Sprintf("device_ids: %v", target.GetDevices().DeviceIds)
	}
	return "<unknown>"
}

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func newTimeoutCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return ctx, cancel
}

var animals = [...]string{
	"aardvark",
	"albatross",
	"alligator",
	"alpaca",
	"ant",
	"anteater",
	"antelope",
	"ape",
	"armadillo",
	"baboon",
	"badger",
	"barracuda",
	"bat",
	"bear",
	"beaver",
	"bee",
	"bird",
	"aves",
	"bison",
	"boar",
	"buffalo",
	"camel",
	"caribou",
	"cassowary",
	"cat",
	"caterpillar",
	"cattle",
	"chamois",
	"cheetah",
	"chicken",
	"chimpanzee",
	"chinchilla",
	"chough",
	"coati",
	"cobra",
	"cockroach",
	"cod",
	"cormorant",
	"coyote",
	"crab",
	"crane",
	"crocodile",
	"crow",
	"curlew",
	"deer",
	"dinosaur",
	"dog",
	"dogfish",
	"dolphin",
	"donkey",
	"dotterel",
	"dove",
	"dragonfly",
	"duck",
	"dugong",
	"dunlin",
	"eagle",
	"echidna",
	"eel",
	"eland",
	"elephant",
	"elephant seal",
	"elk",
	"emu",
	"falcon",
	"ferret",
	"finch",
	"fish",
	"flamingo",
	"fly",
	"fox",
	"frog",
	"gaur",
	"gazelle",
	"gerbil",
	"giant panda",
	"giraffe",
	"gnat",
	"gnu",
	"goat",
	"goldfinch",
	"goosander",
	"goose",
	"gorilla",
	"goshawk",
	"grasshopper",
	"grouse",
	"guanaco",
	"guinea fowl",
	"guinea pig",
	"gull",
	"hamster",
	"hare",
	"hawk",
	"hedgehog",
	"heron",
	"herring",
	"hippo",
	"hornet",
	"horse",
	"hummingbird",
	"hyena",
	"ibex",
	"ibis",
	"jackal",
	"jaguar",
	"jay",
	"jellyfish",
	"kangaroo",
	"kinkajou",
	"koala",
	"komodo dragon",
	"kouprey",
	"kudu",
	"lapwing",
	"lark",
	"lemur",
	"leopard",
	"lion",
	"llama",
	"lobster",
	"locust",
	"loris",
	"louse",
	"lyrebird",
	"magpie",
	"mallard",
	"mammoth",
	"manatee",
	"mandrill",
	"mink",
	"mole",
	"mongoose",
	"monkey",
	"moose",
	"mouse",
	"mosquito",
	"narwhal",
	"newt",
	"nightingale",
	"octopus",
	"okapi",
	"opossum",
	"ostrich",
	"otter",
	"owl",
	"oyster",
	"panther",
	"parrot",
	"panda",
	"partridge",
	"peafowl",
	"pelican",
	"penguin",
	"pheasant",
	"pig",
	"pigeon",
	"polar bear",
	"pony",
	"porcupine",
	"porpoise",
	"prairie dog",
	"quail",
	"quelea",
	"quetzal",
	"rabbit",
	"raccoon",
	"ram",
	"rat",
	"raven",
	"red deer",
	"red panda",
	"reindeer",
	"rhinoceros",
	"rook",
	"salamander",
	"salmon",
	"sand dollar",
	"sandpiper",
	"sardine",
	"sea lion",
	"sea urchin",
	"seahorse",
	"seal",
	"shark",
	"sheep",
	"shrew",
	"skunk",
	"sloth",
	"snail",
	"snake",
	"spider",
	"squirrel",
	"starling",
	"stegosaurus",
	"swan",
	"tapir",
	"tarsier",
	"termite",
	"tiger",
	"toad",
	"turkey",
	"turtle",
	"vicuña",
	"wallaby",
	"walrus",
	"wasp",
	"water buffalo",
	"weasel",
	"whale",
	"wolf",
	"wolverine",
	"wombat",
	"wren",
	"yak",
	"zebra",
}
