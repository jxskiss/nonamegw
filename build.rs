fn main() {
    build_prost();
}

fn build_prost() {
    prost_build::Config::new()
        .out_dir("comet/src/proto")
        .compile_protos(
            &[
                "proto/packet.proto",
                "proto/protocol.proto",
                "proto/messag.proto",
                "proto/data.proto",
                "proto/cometsvc.proto",
            ],
            &["proto/"],
        )
        .unwrap();
}
