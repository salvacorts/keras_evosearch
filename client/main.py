import time
import grpc

import tensorflow as tf
from tensorflow import keras
from keras.models import Sequential
from keras.layers import Dense, Dropout
from keras.metrics import Recall
import numpy as np
from sklearn import preprocessing
from sklearn.model_selection import train_test_split
import pandas as pd
from sklearn.preprocessing import LabelEncoder
from protobuf.api_pb2 import ModelResults, Empty
from protobuf.api_pb2_grpc import APIStub

with grpc.insecure_channel("localhost:10000") as channel:
    stub = APIStub(channel)

    df = pd.read_csv('https://raw.githubusercontent.com/jeffheaton/proben1/master/cancer/breast-cancer-wisconsin.data',header=None)
    df.drop(columns=[0], inplace=True)
    df.replace('?', -99999, inplace=True)
    df[9] = df[9].map(lambda x: 1 if x == 4 else 0)

    X = np.array(df.drop([9], axis=1))
    y = np.array(df[9])

    scaler = preprocessing.MinMaxScaler()
    X = scaler.fit_transform(X)
    labelencoder_Y = LabelEncoder()
    y = labelencoder_Y.fit_transform(y)
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

    while True:
        params = stub.GetModelParams(Empty())
        print(params)
        optimizer=params.optimizer
        activacion=params.activation_func
        learningrate=params.learning_rate
        num_units= params.layers
        dropout_rate=0.25

        def createmodel():
            global model
            model = Sequential()
            model.add(Dense(units=num_units[0], input_shape=(9,) ) )
            
            for layer in num_units:
                model.add( Dropout(dropout_rate) )
                model.add( Dense(units=layer, activation=activacion) )
            
            model.add( Dropout(dropout_rate) )
            model.add( Dense(units=layer, activation=activacion) )
            
            model.add( Dropout(dropout_rate) )
            model.add( Dense(1,activation=activacion) )
            model.compile(optimizer=optimizer(learningrate), loss=keras.losses.binary_crossentropy, metrics=[Recall()])

            model.summary()
    
    
        def trainmodel():
            global model
            model.fit(X_train,y_train,batch_size=32,epochs=50,verbose=1, validation_data=(X_test, y_test))

        createmodel()
        trainmodel()
        loss, recall = model.evaluate(X_test, y_test, verbose=1, batch_size=32)

        results = ModelResults()
        results.model_id = params.model_id
        results.recall = recall
        _ = stub.ReturnModel(results)

        time.sleep(2)
