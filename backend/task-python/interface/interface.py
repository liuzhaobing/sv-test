# -*- coding:utf-8 -*-
import json
import grpc
from google.protobuf import json_format


def pb_to_json(pb):
    return json_format.MessageToJson(pb)


def json_to_pb(json_obj, message):
    return json_format.Parse(json.dumps(json_obj, indent=4), message)


def yield_message(message):
    yield message


class Interface:
    def __init__(self, url, stub):
        self.channel = grpc.insecure_channel(url)
        self.stub = stub(self.channel)

    def __del__(self):
        self.channel.close()

    @staticmethod
    def call(payload, message, func, iterator=False):
        request = yield_message(json_to_pb(payload, message)) if iterator else json_to_pb(payload, message)
        return list(func(request))
