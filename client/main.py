import tensorflow as tf
from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import Dense, Dropout

from sklearn import preprocessing
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder

import pandas as pd
import numpy as np

import argparse
import time

import grpc
from protobuf.api_pb2 import ModelParameters, ModelResults, Empty, Optimizer, ActivationFunc
from protobuf.api_pb2_grpc import APIStub

def ParseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser()

    parser.add_argument("-s","--server", action="store", dest="server",
                        type=str, default="localhost:10000",
                        help="Server address to connect to")

    return parser.parse_args()

def CreateModel(params: ModelParameters) -> Sequential:
    model = Sequential()

    activation_func = ActivationFunc.Name(params.activation_func).lower()
    optimizers = {
        Optimizer.Adam: tf.keras.optimizers.Adam,
        Optimizer.SGD: tf.keras.optimizers.SGD,
        Optimizer.RMSprop: tf.keras.optimizers.RMSprop
    }
    
    for i, layer in enumerate(params.layers):
        if i == 0:
            model.add(Dense(units=layer.num_neurons,
                            activation=activation_func,
                            input_shape=(9,)))
        else:
            model.add(Dense(units=layer.num_neurons,
                            activation=activation_func))

        if params.dropout:
            model.add(Dropout(0.25))
    
    model.add(Dense(1, activation="sigmoid"))

    optimizer = optimizers[params.optimizer]
    model.compile(optimizer=optimizer(params.learning_rate),
                  loss=tf.keras.losses.binary_crossentropy,
                  metrics=[tf.keras.metrics.Recall()])

    return model

if __name__ == "__main__":
    args = ParseArgs()

    df = pd.read_csv('https://raw.githubusercontent.com/jeffheaton/proben1/master/cancer/breast-cancer-wisconsin.data', header=None)
    df.drop(columns=[0], inplace=True)
    df.replace('?', -99999, inplace=True)
    df[9] = df[9].map(lambda x: 1 if x == 4 else 0)

    X = np.array(df.drop([9], axis=1))
    y = np.array(df[9])

    scaler = preprocessing.MinMaxScaler()
    X = scaler.fit_transform(X)
    labelencoder_Y = LabelEncoder()
    y = labelencoder_Y.fit_transform(y)

    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2,random_state=0)
    X_train, X_val, y_train, y_val = train_test_split(X_train, y_train, test_size=0.25, random_state=0) 

    print(f"Connecting to {args.server}")
    with grpc.insecure_channel(args.server) as channel:
        stub = APIStub(channel)

        while True:
            try:
                params = stub.GetModelParams(Empty())
                print(params)

                model = CreateModel(params)
                
                # Train model
                model.fit(X_train, y_train,
                          batch_size=32,
                          epochs=10,
                          verbose=1,
                          validation_data=(X_val, y_val)) 
            
                loss, recall = model.evaluate(X_test, y_test,
                                              verbose=0,
                                              batch_size=32)

                results = ModelResults()
                results.model_id = params.model_id
                results.recall = recall

                print(f"Returning params")
                _ = stub.ReturnModel(results)
            except grpc.RpcError as rpc_error:
                if rpc_error.code() == grpc.StatusCode.CANCELLED:
                    print("No models to evaluate now. Sleeping...")
                    time.sleep(0.5)
                elif rpc_error.code() == grpc.StatusCode.UNAVAILABLE:
                    print("Server is down")
                    exit(0)
                else:
                    print(rpc_error)
                    exit(1)
                           
