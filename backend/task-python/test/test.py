import requests

from badcase_tagging import CMSBadCase
from utils.utils_handler import Handlers

"""
临时用于收集整理线上日志
"""


class RealData(CMSBadCase):
    def get_all_data_count(self, start_time, end_time):
        """统计该时段内满足条件的数据总条数"""
        payload = {
            "[]": {
                "dwm_svo_anno_label_event_i_d": {
                    # "label_type_id{}": [
                    #     '3',
                    #     '8'
                    # ],
                    "nlu_event_time&{}": f">='{start_time}',<='{end_time}'",
                    # "domain__domain_name": "music",
                    # "qa_from": "system_service"
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
                    # "label_type_id{}": [
                    #     '3',
                    #     '8'
                    # ],
                    # "qa_from": "system_service",
                    # "domain__domain_name": "music",
                    "nlu_event_time&{}": f">='{start_time}',<='{end_time}'",
                    "@column": "question_id,question_text,qa_from,domain__domain_name,intent__intent_name,sv_answer_text"
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
                try:
                    mp = {
                        "id": 0,
                        "question": d["dwm_svo_anno_label_event_i_d"]["question_text"],
                        "source": d["dwm_svo_anno_label_event_i_d"]["qa_from"],
                        "domain": d["dwm_svo_anno_label_event_i_d"]["domain__domain_name"],
                        "intent": d["dwm_svo_anno_label_event_i_d"]["intent__intent_name"],
                        "question_id": d["dwm_svo_anno_label_event_i_d"]["question_id"],
                        "sv_answer_text": d["dwm_svo_anno_label_event_i_d"]["sv_answer_text"]
                    }
                    self.weekly_data_old.append(mp)
                    if exclude_domain and mp["domain"] in exclude_domain:
                        self.weekly_data_old.remove(mp)
                    if exclude_intent and mp["intent"] in exclude_intent:
                        self.weekly_data_old.remove(mp)
                except:
                    pass
        return self.weekly_data_old


if __name__ == '__main__':
    instance = RealData()
    date = [
        "2022-01-01",
        "2022-02-01",
        "2022-03-01",
        "2022-04-01",
        "2022-05-01",
        "2022-06-01",
        "2022-07-01",
        "2022-08-01",
        "2022-09-01",
        "2022-10-01",
        "2022-11-01",
        "2022-12-01",
        "2023-01-01",
        "2023-02-01"
    ]
    data = []
    for i in range(len(date) - 1):
        start = date[i]
        end = date[i + 1]
        print(start, end)
        this_batch_data = instance.get_all_data(start, end)
        data += this_batch_data
    new_data = Handlers.list_map_count_and_sort(data, "question")
    Handlers.write_list_map_as_excel(new_data,
                                     excel_writer=r"D:\线上日志.xlsx",
                                     sheet_name="Sheet1",
                                     index=False)
