FROM python:latest

RUN apt-get -y update \
    && apt-get install -y ffmpeg  \
    && rm -rf /var/lib/apt/lists/*
RUN mkdir /srv/data

ADD ytd2feed/app /srv
COPY ytd2feed/app/ /srv/app/
ADD ytd2feed/main.py /srv 

ADD ytd2feed/requirements.txt /srv

WORKDIR /srv
RUN pip install -r requirements.txt

CMD ["python", "-u", "/srv/main.py"]
