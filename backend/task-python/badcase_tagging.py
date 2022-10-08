# -*- coding:utf-8 -*-
"""
1.每周二晚 先从线上拉取上个周符合条件的数据集
2.将数据集进行聚合 形成出现频次统计
3.将数据集按照出现频次倒序排列
4.数据集与用例集去重
5.去重后的数据集 按照倒序排列 节选前1000条频次高的数据作为本周新用例集
6.1000条用例进行sv端到端测试 输出结果给到标注同学
7.标注完成
8.用例入库
"""
import requests
import datetime
from datetime import timedelta

from utils.utils_handler import Handlers
from utils.utils_mysql import DataBaseMySQL


class CMSBadCase:
    """每周二晚上 从线上拉取badcase 用于自动化测试集建设"""

    def __init__(self):
        self.max_version = None
        self.max_id = None
        self.weekly_data_new = None
        self.weekly_data_old = []
        self.count = 0
        self.pagesize = 100
        self.url = "http://172.16.23.83:30515/roc/quoto/cdmCo"
        self.headers = {
            'cookie': "JSESSIONID=98A2B1C17DC869F1E9EEFAA46A665B91",
            'Content-Type': "application/json",
            'token': "UUNWAX0Z2APUBZYJ9S1F"
        }

    def get_all_data_count(self, start_time, end_time):
        """统计该时段内满足条件的数据总条数"""
        payload = {
            "[]": {
                "dwm_svo_anno_label_event_i_d": {
                    "label_type_id{}": [
                        '3',
                        '8'
                    ],
                    "nlu_event_time&{}": f">='{start_time}',<='{end_time}'",
                    "qa_from": "system_service"
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
                        '3',
                        '8'
                    ],
                    "qa_from": "system_service",
                    "nlu_event_time&{}": f">='{start_time}',<='{end_time}'",
                    "@column": "question_id,question_text,qa_from,domain__domain_name,intent__intent_name"
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

        for p in range(pages):
            data = self.get_data_by_page(start_time, end_time, p)
            for d in data:
                mp = {
                    "id": 0,
                    "question": d["dwm_svo_anno_label_event_i_d"]["question_text"],
                    "source": d["dwm_svo_anno_label_event_i_d"]["qa_from"],
                    "domain": d["dwm_svo_anno_label_event_i_d"]["domain__domain_name"],
                    "intent": d["dwm_svo_anno_label_event_i_d"]["intent__intent_name"],
                    "question_id": d["dwm_svo_anno_label_event_i_d"]["question_id"]
                }
                self.weekly_data_old.append(mp)
                if exclude_domain and mp["domain"] in exclude_domain:
                    self.weekly_data_old.remove(mp)
                if exclude_intent and mp["intent"] in exclude_intent:
                    self.weekly_data_old.remove(mp)
        return self.weekly_data_old

    def sort_and_duplicate_data(self):
        """根据现有用例 对新来的数据 统计 排序 去重处理"""
        dbinfo1 = {
            'host': '172.16.23.33',
            'user': 'root',
            'password': '',
            'port': 3306,
            'dbname': 'nlpautotest'
        }
        database_cases = DataBaseMySQL(dbinfo1).query("select * from skill_base_test;")
        info = DataBaseMySQL(dbinfo1).query("select max(id) id, max(case_version) case_version from skill_base_test;")
        self.max_id = info[0]["id"]
        self.max_version = info[0]["case_version"]

        new_data = Handlers.list_map_count_and_sort(self.weekly_data_old, "question")
        new_data = Handlers.list_map_duplicate_by_another_list_map(database_cases, new_data, "question")
        self.weekly_data_new = new_data[:1000]
        return self.weekly_data_new, self.max_id, self.max_version

    def return_suitable_cases(self):
        """配置导入用例库的对应的格式"""
        final_data = []
        id_start = self.max_id // 100 * 100 + 100
        for case in self.weekly_data_new:
            id_start += 1
            case.pop("counter")
            case["usetest"] = 4
            case["case_version"] = self.max_version
            case["id"] = id_start
            final_data.append(case)
        return final_data


class Tagging:
    def __init__(self):
        self.base_url = "http://annotaion.region-dev-1.service.iamidata.com:31123"
        self.username = "adminbz001"
        self.password = "123456"
        self.headers = {"Content-Type": "application/json"}
        self.login()

    def login(self):
        url = "/v1/user/login"
        payload = {"username": self.username, "password": self.password}
        res = requests.request(method="POST", url=self.base_url + url, json=payload, headers=self.headers, verify=False)
        if res.status_code == 200:
            response = res.json()
            token = response["data"]["token"]
            self.headers["Cookie"] = f"jwt={token}"
            self.headers["Authorization"] = token
            return self.headers

    def push_question_id_to_label(self, name, question_ids, env="bj-prod-221"):
        """将需要标注的case推送到标注平台"""
        url = "/v1/label/manual_assign"
        payload = {
            "name": name,
            "env": env,
            "question_ids": question_ids
        }
        res = requests.request(method="POST", url=self.base_url + url, json=payload, headers=self.headers, verify=False)
        return res


def badcase_tagging_push(start_time=None, end_time=None, exclude_domain=None, exclude_intent=None):
    now = datetime.datetime.now()
    if not start_time:
        start_time = str(now - timedelta(days=now.weekday() + 7))[:10] + " 00:00:00"
    if not end_time:
        end_time = str(now - timedelta(days=now.weekday()))[:10] + " 00:00:00"

    t = CMSBadCase()
    t.get_all_data(start_time, end_time, exclude_domain, exclude_intent)
    t.sort_and_duplicate_data()
    list_map = t.return_suitable_cases()
    task_name = "skill_case_" + str(t.max_version) + "_week" + start_time[:10] + "_" + end_time[5:10]
    task_question_ids = []
    for i in list_map:
        task_question_ids.append(i["question_id"])

    result = Tagging().push_question_id_to_label(name=task_name, question_ids=task_question_ids, env="bj-prod-221")
    return result, task_name


if __name__ == '__main__':
    badcase_tagging_push(exclude_domain=["indoornavigation", "around"])
