import time
import grpc
from protobuf.api_pb2 import ModelResults, Empty
from protobuf.api_pb2_grpc import APIStub

with grpc.insecure_channel("localhost:10000") as channel:
    stub = APIStub(channel)

    while True:
        params = stub.GetModelParams(Empty())
        print(params)

        results = ModelResults()
        results.model_id = params.model_id
        results.recall = 1.0
        _ = stub.ReturnModel(results)

        time.sleep(2)
