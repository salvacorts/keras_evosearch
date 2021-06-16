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

from protobuf.api_pb2 import ModelParameters, ModelResults, Empty
from protobuf.api_pb2_grpc import APIStub

def ParseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser()

    parser.add_argument("-s","--server", action="store", dest="server",
                        type=str, default="localhost:10000",
                        help="Server address to connect to")

    return parser.parse_args()

def CreateModel(params: ModelParameters) -> Sequential:
    model = Sequential()
    
    # TODO: map activation func to actual activation func
    for i, layer in enumerate(params.layers):
        if i == 0:
            model.add(Dense(units=layer.num_neurons,
                            activation=params.activation_func,
                            input_shape=(9,)))
        else:
            model.add(Dense(units=layer.num_neurons,
                            activation=params.activation_func))

        if params.dropout:
            model.add(Dropout(0.25))
    
    model.add(Dense(1, activation="sigmoid"))

    # TODO: Map optimizer to actual optimizer
    model.compile(optimizer=params.optimizer(params.learning_rate),
                  loss=tf.keras.losses.binary_crossentropy,
                  metrics=[tf.keras.metrics.Recall()])

    return model

if __name__ == "__main__":
    args = ParseArgs()

    print(f"Connecting to {args.server}")
    with grpc.insecure_channel(args.server) as channel:
        stub = APIStub(channel)

        df = pd.read_csv('https://raw.githubusercontent.com/jeffheaton/proben1/master/cancer/breast-cancer-wisconsin.data', header=None)
        df.drop(columns=[0], inplace=True)
        df.replace('?', np.nan, inplace=True) # TODO: Why? - Porque '?' no es un número y da errores 
        df[9] = df[9].map(lambda x: 1 if x == 4 else 0)

        X = np.array(df.drop([9], axis=1))
        y = np.array(df[9])

        # TODO: Set seed?
        scaler = preprocessing.MinMaxScaler()
        X = scaler.fit_transform(X)
        labelencoder_Y = LabelEncoder()
        y = labelencoder_Y.fit_transform(y)

        # TODO: - Set seed - Hecho
        #       - Divide train, validation and test - Hecho
        X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2,random_state=0)
        X_train, X_val, y_train, y_val = train_test_split(X_train, y_train, test_size=0.25, random_state=0) 
      
        while True:
            try:
                params = stub.GetModelParams(Empty())
                print(params)

                model = CreateModel(params)
                
                # Train model
                model.fit(X_train, y_train,
                          batch_size=32,
                          epochs=50,
                          verbose=1,
                          validation_data=(X_val, y_val)) 
            
                loss, recall = model.evaluate(X_test, y_test,
                                            verbose=1,
                                            batch_size=32)

                results = ModelResults()
                results.model_id = params.model_id
                results.recall = recall
                _ = stub.ReturnModel(results)
            except:
                time.sleep(5)


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

            
