# -*- coding:utf-8 -*-
import uuid

import grpc
import talk_pb2
import talk_pb2_grpc


def run():
    def talk_req():
        yield talk_pb2.TalkRequest(is_full=True,
                                   agent_id=666,
                                   session_id=str(uuid.uuid4()),
                                   question_id=str(uuid.uuid4()),
                                   event_type=0,
                                   robot_id="123",
                                   tenant_code="cloudminds",
                                   version="v3",
                                   test_mode=False,
                                   asr=talk_pb2.Asr(lang="CH", text="现在几点了"))

    with grpc.insecure_channel('172.16.23.85:30811') as channel:
        stub = talk_pb2_grpc.TalkStub(channel)
        responses = stub.StreamingTalk(talk_req().__iter__())
        response = list(responses)

    for r in response:
        source = r.source
        print(source)
        cost = r.cost
        print(cost)
        tts = r.tts
        for t in tts:
            print(t.text)
        domain = r.hit_log.fields["domain"].string_value
        print(domain)


if __name__ == '__main__':
    run()
