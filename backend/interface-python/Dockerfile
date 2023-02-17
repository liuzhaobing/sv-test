FROM python:3.9
COPY . /app

WORKDIR /app

ENV LC_ALL C.UTF-8
ENV LANG C.UTF-8
ENV PIPURL "https://pypi.tuna.tsinghua.edu.cn/simple"

RUN pip3 install  -i ${PIPURL} --upgrade pip
RUN pip3 install  -i ${PIPURL} -r requirements.txt

CMD ["python", "app.py"]
