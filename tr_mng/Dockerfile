FROM python:3.13

RUN mkdir /srv/data

ADD tr_mng/app /srv
COPY tr_mng/app/ /srv/app/
ADD tr_mng/main.py /srv

ADD tr_mng/requirements.txt /srv

WORKDIR /srv
RUN pip install -r requirements.txt

CMD ["python", "-u", "/srv/main.py"]