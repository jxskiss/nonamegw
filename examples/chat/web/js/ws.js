class Client {
	constructor(endpoint) {
		this.endpoint = endpoint;
		this.seq = 1;
		this.ready = false;
		this.ws = null;
		this.pending = {};
		this.handler = {};
	}

	connect() {
		var self = this;
		if (this.ws != null) {
			return this.ws
		}
		return this.ws = new Promise(function(resolve, reject) {
            m.request({
                method: "GET",
                url: self.endpoint,
                data: {aid: 1, did: 345, ver: 123123123}
            }).then(function(result) {
                console.log(result);

                var wsEndpoint = "ws://" + result.addresses[0] + "/ws?tok=" + result.token + "&a=1&b=2&c=3";
                var ws = new WebSocket(wsEndpoint, "v2.json");
                var pending = true;
                ws.onerror = function(err) {
                    if (pending) {
                        pending = false;
                        reject(err)
                        return
                    }

                    console.warn("websocket lifetime error:" + err)
                    Object.keys(self.pending).forEach(function(k) {
                        self.pending[k].reject(err);
                        delete self.pending[k];
                    })
                };
                ws.onopen = function() {
                    if (pending) {
                        pending = false
                        resolve(ws)
                    }
                };
                ws.onmessage = function(s) {
                    console.log("received:", s);
                    var packet, msg, request, handler;
                    try {
                        // json protocol
                        packet = JSON.parse(s.data);
                        msg = JSON.parse(atob(packet.payload));
                    } catch (err) {
                        console.warn("parse incoming message error:", err);
                        return
                    }

                    // Notice
                    if (msg.id == void 0 || msg.id == 0) {
                        if (handler = self.handler[msg.method]) {
                            handler.forEach((h) => h(msg.params, self));
                            return
                        }
                        console.warn("no handler for method:", msg.method);
                        return
                    }

                    request = self.pending[msg.id];
                    if (request == null) {
                        console.warn("no pending request for:", msg.method, msg.id);
                        return
                    }

                    delete self.pending[msg.id];
                    if (msg.error != null) {
                        request.reject(msg.error);
                    } else {
                        request.resolve(msg.result);
                    }

                    return;
                };
            })
		})
	}

	call(method, params) {
		var self = this;
		return this.connect()
			.then(function(conn) {
				var seq = self.seq++;
				var dfd = defer();
				self.pending[seq] = dfd;

				var msg, packet;
				msg = {
					id: seq,
					method: method,
					params: params,
				};
				packet = {
					seq_id: seq,
					log_id: "" + new Date().getTime() + "000000",
					payload_encoding: "base64",
					payload_type: "json",
					payload: btoa(JSON.stringify(msg))
				};

				console.log("sending:", packet, msg, JSON.stringify(packet));
				conn.send(JSON.stringify(packet));

				return dfd.promise;
			})
	}

	handle(method, callback) {
		var list = this.handler[method];
		if (list == null) {
			this.handler[method] = [callback];
			return
		}
		list.push(callback);
	}
}

function defer() {
	var d = {}
	d.promise = new Promise(function(resolve, reject) {
		d.resolve = resolve;
		d.reject = reject;
	})
	return d
}
