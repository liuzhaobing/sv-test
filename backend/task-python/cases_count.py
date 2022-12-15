# -*- coding:utf-8 -*-
"""
每日用例统计
"""
import datetime
import json

import requests

from utils.utils_mysql import DataBaseMySQL


class GroupCases:

    def __init__(self):
        self.today_now = datetime.datetime.now()
        self.yesterday_now = str(self.today_now - datetime.timedelta(days=1))[:19]  # 2022-12-14 11:10:23

        with open("conf/config.json", "r") as f:
            config = json.load(f)
        dbinfo1 = config["DATABASE"]["AUTOTEST"]
        self.mysql_172_16_23_33 = DataBaseMySQL(dbinfo1)

    def skill_case_total(self):
        info = self.mysql_172_16_23_33.query("select count(1) count from skill_base_test;")
        return info[0]["count"]

    def skill_case_group_by(self):
        info = self.mysql_172_16_23_33.query(f"select domain,count(1) count "
                                             f"from skill_base_test "
                                             f"group by domain "
                                             f"order by count desc;")
        return info

    def skill_case_new_total(self):
        info = self.mysql_172_16_23_33.query(f"select count(1) count from skill_base_test "
                                             f"where create_time >= '{self.yesterday_now}';")
        return info[0]["count"]

    def skill_case_new_group_by(self):
        info = self.mysql_172_16_23_33.query(f"select domain,count(1) count "
                                             f"from skill_base_test "
                                             f"where create_time >= '{self.yesterday_now}' "
                                             f"group by domain "
                                             f"order by count desc;")
        return info

    def qa_case_total(self):
        info = self.mysql_172_16_23_33.query("select count(1) count from qa_base_test;")
        return info[0]["count"]

    def qa_case_new_total(self):
        info = self.mysql_172_16_23_33.query(f"select count(1) count from qa_base_test "
                                             f"where create_time >= '{self.yesterday_now}';")
        return info[0]["count"]

    def asr_case_total(self):
        info = self.mysql_172_16_23_33.query("select count(1) count from asr_base_test;")
        return info[0]["count"]

    def asr_case_new_total(self):
        info = self.mysql_172_16_23_33.query(f"select count(1) count from asr_base_test "
                                             f"where create_time >= '{self.yesterday_now}';")
        return info[0]["count"]

    def tts_case_total(self):
        info = self.mysql_172_16_23_33.query("select count(1) count from tts_base_test;")
        return info[0]["count"]

    def tts_case_new_total(self):
        info = self.mysql_172_16_23_33.query(f"select count(1) count from tts_base_test "
                                             f"where create_time >= '{self.yesterday_now}';")
        return info[0]["count"]


def get_cases_info_text():
    instance = GroupCases()
    skill_total = instance.skill_case_total()
    qa_total = instance.qa_case_total()
    asr_total = instance.asr_case_total()
    tts_total = instance.tts_case_total()
    total = skill_total + qa_total + asr_total + tts_total

    new_skill_total = instance.skill_case_new_total()
    new_qa_total = instance.qa_case_new_total()
    new_asr_total = instance.asr_case_new_total()
    new_tts_total = instance.tts_case_new_total()
    new_total = new_skill_total + new_qa_total + new_asr_total + new_tts_total

    tmp = f"""测试集建设工作
    目前测试集共计{total}条，其中技能{skill_total}条，QA{qa_total}条，ASR{asr_total}条，TTS{tts_total}条
    昨日新增共计{new_total}条，其中技能条{new_skill_total}条，QA{new_qa_total}条，ASR{new_asr_total}条，TTS{new_tts_total}条\n"""

    if new_skill_total:
        tmp += """技能测试集新增用例具体分布\n"""
        new_skill_group_by = instance.skill_case_new_group_by()
        for item in new_skill_group_by:
            tmp += f"""    {item["domain"]}:{item["count"]}\n"""
    else:
        tmp += """技能测试集具体分布\n"""
        skill_group_by = instance.skill_case_group_by()
        for item in skill_group_by:
            tmp += f"""    {item["domain"]}:{item["count"]}\n"""
    return tmp


def cases_construction():
    requests.request(method="POST",
                     url="https://open.feishu.cn/open-apis/bot/v2/hook/c0ea24df-4894-4aeb-a9df-812b6653564d",
                     json={
                         "msg_type": "text",
                         "content": json.dumps({"text": f"{get_cases_info_text()}"}, ensure_ascii=False)
                     })
