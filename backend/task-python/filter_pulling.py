# -*- coding:utf-8 -*-

"""
定时拉取线上标记为噪声的case
"""
import datetime
import uuid
from datetime import timedelta
import json

import requests
import redis

from badcase_tagging import CMSBadCase
from utils.utils_mysql import DataBaseMySQL

with open("conf/config.json", "r") as f:
    config = json.load(f)


def big_data_tagging_user_table(query):
    """从大数据库中查数据库smartomp_rbac"""
    big_data_dbinfo = config["DATABASE"]["CMS"]
    big_data_dbinfo["dbname"] = "smartomp_rbac"
    return DataBaseMySQL(big_data_dbinfo).query(query)


def find_tagging_user():
    """从数据库中查找标注人员信息"""
    tagging_user_list = big_data_tagging_user_table("select id,username,description from sunny_editor;")
    tagging_user_map = {}
    for u in tagging_user_list:
        tagging_user_map[str(u["id"])] = f'{u["username"]}{u["description"]}'
    return tagging_user_map


class CMSAsrFilter(CMSBadCase):
    def get_all_data_count(self, start_time, end_time):
        """统计该时段内满足条件的数据总条数"""
        payload = {
            "[]": {
                "dwm_svo_anno_label_event_i_d": {
                    "label_type_id{}": [
                        '5'
                    ],
                    "submit_time&{}": f">='{start_time}',<='{end_time}'",
                    "qa_from{}": [
                        'system_service',
                        'common_sense_qa',
                        'openkg_qa',
                        'third_chitchat',
                        'user_default_qa'
                    ]
                },
                "query": 1,
            },
            "total@": "/[]/total"
        }
        res = requests.request(method="POST", url=self.url, headers=self.headers, json=payload)
        self.count = res.json()["total"]
        return self.count

    def get_data_by_page(self, start_time, end_time, page):
        """从大数据这边分页查询满足条件的数据"""
        payload = {
            "[]": {
                "dwm_svo_anno_label_event_i_d": {
                    "label_type_id{}": [
                        '5'
                    ],
                    "qa_from{}": [
                        'system_service',
                        'common_sense_qa',
                        'openkg_qa',
                        'third_chitchat',
                        'user_default_qa'
                    ],
                    "submit_time&{}": f">='{start_time}',<='{end_time}'",
                    "@column": "question_id,question_text,qa_from,domain__domain_name,intent__intent_name,param_info,label_type_name,submit_time,event_time,operator_id,sv_answer_text,robot__robot_type_name,sv_agent_id"
                },
                "page": page,
                "count": self.pagesize
            }
        }
        res = requests.request(method="POST", url=self.url, headers=self.headers, json=payload).json()
        return res["[]"]

    def get_all_data(self, start_time, end_time, exclude_domain=None, exclude_intent=None):
        """用分页查询 查询所有页的数据 并配置指定格式的list[map]"""
        self.count = self.get_all_data_count(start_time, end_time)
        pages = self.count // self.pagesize
        if self.count % self.pagesize != 0:
            pages += 1

        tagging_user_map = find_tagging_user()

        for p in range(pages):
            data = self.get_data_by_page(start_time, end_time, p)
            for d in data:
                mp = {
                    "question": d["dwm_svo_anno_label_event_i_d"]["question_text"],
                    "source": d["dwm_svo_anno_label_event_i_d"]["qa_from"],
                    "domain": d["dwm_svo_anno_label_event_i_d"]["domain__domain_name"],
                    "intent": d["dwm_svo_anno_label_event_i_d"]["intent__intent_name"],
                    "sv_answer": d["dwm_svo_anno_label_event_i_d"]["sv_answer_text"],
                    "robot_type": d["dwm_svo_anno_label_event_i_d"]["robot__robot_type_name"],
                    "agent_id": d["dwm_svo_anno_label_event_i_d"]["sv_agent_id"],
                    "label_name": d["dwm_svo_anno_label_event_i_d"]["label_type_name"],
                    "label_time": d["dwm_svo_anno_label_event_i_d"]["submit_time"],
                    "label_operator": tagging_user_map[d["dwm_svo_anno_label_event_i_d"]["operator_id"]],
                    "event_time": d["dwm_svo_anno_label_event_i_d"]["event_time"],
                    "question_id": d["dwm_svo_anno_label_event_i_d"]["question_id"]
                }
                mp = resign_case(mp)
                self.weekly_data_old.append(mp)
        return self.weekly_data_old

    def sort_and_duplicate_data(self):
        """根据现有用例 对新来的数据 统计 排序 去重处理"""
        redis_connect = redis.StrictRedis(host='172.16.23.85', port=31961, db=0)

        def redis_get(name):
            result = redis_connect.get(name)
            if not result:
                return None
            return result.decode()

        def redis_set(name, value):
            result = redis_connect.set(name, value)
            if not result:
                return None
            return result

        already_checked_questions = redis_get("nlpautotest")
        if already_checked_questions:
            already_checked_questions = json.loads(already_checked_questions)
        else:
            already_checked_questions = []

        def list_duplication(list_obj, list_map_obj, column_name):
            for i in range(len(list_map_obj)):
                if list_map_obj[i][column_name] not in list_obj:
                    list_map_obj[i]["show_before"] = "no"
                    list_obj.append(list_map_obj[i][column_name])
                else:
                    list_map_obj[i]["show_before"] = "yes"
            return list_map_obj, list_obj

        self.weekly_data_new, this_time_questions = list_duplication(already_checked_questions, self.weekly_data_old,
                                                                     "question")
        redis_set("nlpautotest", json.dumps(this_time_questions, ensure_ascii=False))
        return self.weekly_data_new


def resign_case(case_info):
    developer = ""
    if case_info["source"] == "openkg_qa":
        developer = "@Bei Chen 陈贝"
    if case_info["source"] in ["user_default_qa", "third_chitchat"]:
        developer = "@Jessica Li 李翠姣"
    if case_info["source"] == "common_sense_qa":
        developer = "@Raino Wu 吴雨浓"

    if case_info["source"] == "system_service":
        developer = "@Xia Fu 付霞"
    if case_info["intent"] == "SingerSong" or case_info["source"] in ["dialogYou", "dialogNow"]:
        developer = "@Zero Zhou 周成浩"
    case_info["bug_owner"] = developer
    return case_info


def asr_filter_pull(start_time=None, end_time=None,
                    server_ip="127.0.0.1:27997",
                    feishu_url="https://open.feishu.cn/open-apis/bot/v2/hook/5320b9cc-480d-4213-aee4-399ead257c19"):
    now = datetime.datetime.now()
    if not start_time:
        start_time = str(now - timedelta(days=1))[:10] + " 00:00:00"
    if not end_time:
        end_time = str(now - timedelta(days=0))[:10] + " 00:00:00"
    t = CMSAsrFilter()
    t.get_all_data(start_time, end_time)
    list_map = t.sort_and_duplicate_data()
    file_name = f"asr_filter_badcase_{start_time[:10]}_{end_time[5:10]}_"
    if not list_map:
        return None

    # 将数据推送到下载服务器中
    response = requests.request(method="POST", url=f"http://{server_ip}/api/v1/export",
                                json={"name": file_name, "data": list_map})
    try:
        file_name = response.json()["data"]["data"]
        download_url = f"http://{server_ip}/api/v1/download?filename={file_name}"

        feishu_text = f"""干扰测试结果需处理{end_time[:10]}：\n"""
        developer_map = {}
        for item in list_map:
            if developer_map.__contains__(item["bug_owner"]):
                developer_map[item["bug_owner"]] += 1
            else:
                developer_map[item["bug_owner"]] = 1
        for k, v in developer_map.items():
            feishu_text += f"    {k}: {v}\n"

        # 写mongo asr_filter_results表
        from pymongo import MongoClient
        client = MongoClient(host="mongodb://root:123456@172.16.23.33:27927/admin?connect=direct")
        database = client["smartest"]
        asr_filter_results = database["asr_filter_results"]
        result1 = asr_filter_results.insert_many(list_map)
        data_length = len(list_map)
        success_length = len(result1.inserted_ids)
        # 写mongo tasks表
        tasks = database["tasks"]
        tasks.insert_one({
            "job_instance_id": str(uuid.uuid4()),
            "task_name": "干扰测试",
            "task_type": "asr_filter",
            "status": 32,
            "progress_percent": 100,
            "progress": f"{success_length}/{data_length}",
            "accuracy": 0,
            "message": feishu_text,
            "start_time": start_time,
            "end_time": end_time,
            "result_file": file_name
        })

        feishu_text += download_url
        requests.request(method="POST",
                         url=feishu_url,
                         json={
                             "msg_type": "text",
                             "content": json.dumps({"text": feishu_text}, ensure_ascii=False)
                         })
    except:
        return None


if __name__ == '__main__':
    asr_filter_pull()
