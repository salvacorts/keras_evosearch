import time
from typing import Counter
import grpc
import argparse
from protobuf.api_pb2 import ModelResults, Empty
from protobuf.api_pb2_grpc import APIStub

def ParseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser()

    parser.add_argument("-s","--server", action="store", dest="server",
                        type=str, default="localhost:10000",
                        help="Server address to connect to")

    return parser.parse_args()


if __name__ == "__main__":
    args = ParseArgs()

    print(f"Connecting to {args.server}")
    with grpc.insecure_channel(args.server) as channel:
        stub = APIStub(channel)

        while True:
            try:
                params = stub.GetModelParams(Empty())
                print(params)

                results = ModelResults()
                results.model_id = params.model_id
                results.recall = 0.8
                _ = stub.ReturnModel(results)

                time.sleep(2)
            except Exception as e:
                #print(e)
                time.sleep(5)

            
