FROM python:3.13

RUN mkdir /srv/data

ADD finance/app /srv
COPY finance/app/ /srv/app/
ADD finance/main.py /srv

ADD finance/requirements.txt /srv

WORKDIR /srv
RUN pip install -r requirements.txt

CMD ["python", "-u", "/srv/main.py"]