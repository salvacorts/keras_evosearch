# docker build -f client/Dockerfile -t kerasevo/client .
FROM tensorflow/tensorflow:latest

COPY client /var/client

WORKDIR /var/client
RUN pip install --upgrade pip
RUN pip install -r requirements.txt

ENTRYPOINT [ "python", "main.py" ]