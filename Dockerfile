FROM python:3.12
WORKDIR /usr/local/app

COPY main.py .
COPY out .
RUN pip install requests

CMD ["python", "main.py"]