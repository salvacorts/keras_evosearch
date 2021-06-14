# docker build -f client/Dockerfile -t kerasevo/client .
FROM python:latest

COPY client /var/client

WORKDIR /var/client
RUN pip install -r requirements.txt

ENTRYPOINT [ "python", "main.py" ]