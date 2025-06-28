FROM python:3
# RUN apt-get update && apt-get install -y

WORKDIR /app

# RUN pip install --upgrade pip
RUN pip install requests

COPY . .
# RUN mkdir out

EXPOSE 8080

CMD ["python", "./main.py"]