# -*- coding:utf-8 -*-
import datetime
import json
import logging
import requests
import pymysql
from pymongo import MongoClient
from apscheduler.schedulers.blocking import BlockingScheduler

logging.basicConfig(level=logging.INFO,
                    format="[%(asctime)s] [%(levelname)s] [%(funcName)s]: %(message)s")


class SmartVoiceM8:
    def __init__(self, host: str = "mongodb://172.16.32.2:31140/sv_log?connect=direct",
                 sv_log_database: str = "sv_log",
                 sv_log_collection: str = "sv_log",
                 hit_log_database: str = "asrctrl",
                 hit_log_collection: str = "hitlog"):
        self.client = MongoClient(host=host)

        self.sv_log_database = self.client[sv_log_database]
        self.sv_log = self.sv_log_database[sv_log_collection]

        self.hit_log_database = self.client[hit_log_database]
        self.hit_log = self.hit_log_database[hit_log_collection]

        date_today = datetime.date.today()
        self.datetime_today = datetime.datetime.combine(date_today, datetime.time(0, 0, 0))
        self.timestamp_today = int(self.datetime_today.timestamp() * 1000)

        date_7_days_ago = date_today - datetime.timedelta(days=7)
        self.datetime_7_days_ago = datetime.datetime.combine(date_7_days_ago, datetime.time(0, 0, 0))
        self.timestamp_7_days_ago = int(self.datetime_7_days_ago.timestamp() * 1000)

    def nlp_access_count(self, start_date=None, end_date=None):
        """NLP访问量
        :param start_date: start_date = datetime.datetime(2023, 3, 22)
        :param end_date: end_date = datetime.datetime(2023, 3, 29)
        :return: int
        """
        filter = {"time": {"$gte": start_date if start_date else self.datetime_7_days_ago,
                           "$lt": end_date if end_date else self.datetime_today}}
        return self.sv_log.count_documents(filter=filter)

    def nlp_cost_avg(self, start_date=None, end_date=None):
        """NLP耗时均值
        :param start_date: start_date = datetime.datetime(2023, 3, 22)
        :param end_date: end_date = datetime.datetime(2023, 3, 29)
        :return: float
        """
        pipline = [
            {"$match": {"time": {"$gte": start_date if start_date else self.datetime_7_days_ago,
                                 "$lt": end_date if end_date else self.datetime_today}}},
            {"$group": {"_id": "null", "costAvg": {"$avg": "$cost"}}},
        ]
        result = self.sv_log.aggregate(pipline)
        for doc in result:
            return doc["costAvg"]

    def nlp_max_qps(self, start_date=None, end_date=None):
        """NLP QPS计算
        :param start_date: start_date = datetime.datetime(2023, 3, 22)
        :param end_date: end_date = datetime.datetime(2023, 3, 29)
        :return: float
        """
        pipline = [
            {"$match": {"time": {"$gte": start_date if start_date else self.datetime_7_days_ago,
                                 "$lt": end_date if end_date else self.datetime_today}}},
            {"$project": {"fmtTime": {"$dateToString": {"format": "%Y%m%d%H%M", "date": "$time"}}}},
            {"$group": {"_id": "$fmtTime", "count": {"$sum": 1}}},
            {"$sort": {"count": -1}}
        ]
        result = self.sv_log.aggregate(pipline)
        for doc in result:
            return doc["count"] / 60

    def asr_access_count(self, start_date=None, end_date=None):
        """
        :param start_date: 十三位时间戳
        :param end_date:  十三位时间戳
        :return:
        """
        return self.hit_log.count_documents({"asr_start": {
            "$gt": f"{start_date if start_date else self.timestamp_7_days_ago}",
            "$lt": f"{end_date if end_date else self.timestamp_today}"}})

    def asr_cost_avg(self, start_date=None, end_date=None):
        """
        :param start_date: 十三位时间戳
        :param end_date:  十三位时间戳
        :return:
        """
        pipline = [
            {"$match": {"asr_start": {
                "$gt": f"{start_date if start_date else self.timestamp_7_days_ago}",
                "$lt": f"{end_date if end_date else self.timestamp_today}"}}},
            {"$group": {"_id": "null", "asrTotalAvg": {"$avg": "$asr_total"}, "gapAvg": {"$avg": "$gap"}}}
        ]
        result = self.hit_log.aggregate(pipline)
        for doc in result:
            return doc["asrTotalAvg"]

    def asr_max_qps(self, start_date=None, end_date=None):
        """
        :param start_date: 十三位时间戳
        :param end_date:  十三位时间戳
        :return:
        """
        pipline = [
            {"$match": {"asr_start": {
                "$gt": f"{start_date if start_date else self.timestamp_7_days_ago}",
                "$lt": f"{end_date if end_date else self.timestamp_today}"}}},
            {"$project": {"minute": {"$substrBytes": ["$asr_start", 0, 8]}}},
            {"$group": {"_id": "$minute", "count": {"$sum": 1}}},
            {"$sort": {"count": -1}}
        ]
        result = self.hit_log.aggregate(pipline)
        for doc in result:
            return doc["count"] / 100


class DataBaseMySQL:
    def __init__(self, db_info):
        self.conn = pymysql.connect(host=db_info['host'], port=db_info['port'], user=db_info['user'],
                                    password=db_info['password'], db=db_info['dbname'],
                                    charset='utf8', autocommit=True)
        self.cursor = self.conn.cursor(cursor=pymysql.cursors.DictCursor)  # 读取为列表+字典格式

    def query(self, query_string):
        self.cursor.execute(query_string)
        return self.cursor.fetchall()

    def __del__(self):
        self.cursor.close()
        self.conn.close()


class Smartest:
    def __init__(self,
                 host: str = "mongodb://root:123456@172.16.23.33:27927/admin?connect=direct",
                 database: str = "smartest",
                 tasks: str = "tasks",
                 skill: str = "skill_results",
                 qa: str = "qa_results"):
        self.db_info = {
            "host": "172.16.23.33",
            "user": "root",
            "password": "",
            "port": 3306,
            "dbname": "nlpautotest",
            "dbtype": "mysql",
            "dbengine": "pymysql"
        }
        self.mysql_172_16_23_33 = DataBaseMySQL(self.db_info)
        self.client = MongoClient(host=host)

        self.smartest_database = self.client[database]
        self.smartest_task = self.smartest_database[tasks]
        self.smartest_skill = self.smartest_database[skill]
        self.smartest_qa = self.smartest_database[qa]

    def skill_case_count(self):
        info = self.mysql_172_16_23_33.query("select count(1) count from skill_base_test;")
        return info[0]["count"]

    def qa_case_count(self):
        info = self.mysql_172_16_23_33.query("select count(1) count from qa_base_test;")
        return info[0]["count"]

    def qa_last_task_info(self):
        pipline = [
            {
                "$match": {
                    "task_name": "每日QA测试线上环境",
                    "status": 32
                }
            },
            {
                "$sort": {
                    "start_time": - 1
                }
            },
            {
                "$limit": 1
            }
        ]
        result = self.smartest_task.aggregate(pipline)
        job_instance_id = ""
        for doc in result:
            job_instance_id = doc["job_instance_id"]

        case_total = self.smartest_qa.count_documents({"job_instance_id": job_instance_id})
        failed_total = self.smartest_qa.count_documents({"job_instance_id": job_instance_id, "is_pass": False})
        return case_total, failed_total

    def skill_last_task_info(self):
        pipline = [
            {
                "$match": {
                    "task_name": "每日技能测试FIT环境",
                    "status": 32
                }
            },
            {
                "$sort": {
                    "start_time": - 1
                }
            },
            {
                "$limit": 1
            }
        ]
        result = self.smartest_task.aggregate(pipline)
        job_instance_id = ""
        for doc in result:
            job_instance_id = doc["job_instance_id"]

        max_case_version = self.smartest_skill.aggregate([
            {"$match": {"job_instance_id": job_instance_id}},
            {"$group": {"_id": "null", "max": {"$max": "$case_version"}}}])
        max_version = 0
        for doc in max_case_version:
            max_version = doc["max"]

        second_max_version = max_version - 1
        max_version_case_total = self.smartest_skill.count_documents(
            {"job_instance_id": job_instance_id})
        max_version_failed_total = self.smartest_skill.count_documents(
            {"job_instance_id": job_instance_id, "is_pass": False})
        second_max_version_case_total = self.smartest_skill.count_documents(
            {"job_instance_id": job_instance_id, "case_version": {"$lt": max_version}})
        second_max_version_failed_total = self.smartest_skill.count_documents(
            {"job_instance_id": job_instance_id, "case_version": {"$lt": max_version}, "is_pass": False})

        return [
            {
                "case_version": max_version,
                "case_total": max_version_case_total,
                "case_failed": max_version_failed_total
            },
            {
                "case_version": second_max_version,
                "case_total": second_max_version_case_total,
                "case_failed": second_max_version_failed_total
            }
        ]


SMARTEST = """
自动化测试
    共计：{run_total}，待解决：{run_failed}
    技能{max_version}版本：{max_version_case_total}，待解决：{max_version_failed_total}
    技能{second_max_version}版本：{second_max_version_case_total}，待解决：{second_max_version_failed_total}
    QA：{qa_total}，待解决：{qa_failed}
    
测试集建设
    自动化测试技能共计：{case_total_skill}
    自动化测试QA共计：{case_total_qa}
""".strip()

ASR_TEMPLATE = """
ASR Ctrl
    共计访问量：{access_total}，最大并发：{max_qps}，平均时延：{cost_avg}
""".strip()

NLP_TEMPLATE = """
NLP服务器
    共计访问量：{access_total}，最大并发：{max_qps}，平均时延：{cost_avg}
""".strip()


def main():
    instance = Smartest()
    qa_total, qa_failed = instance.qa_last_task_info()
    skill_results = instance.skill_last_task_info()
    smartest_result = SMARTEST.format(
        case_total_qa=instance.qa_case_count(),
        case_total_skill=instance.skill_case_count(),
        qa_total=qa_total,
        qa_failed=qa_failed,
        second_max_version=skill_results[-1]["case_version"],
        second_max_version_case_total=skill_results[-1]["case_total"],
        second_max_version_failed_total=skill_results[-1]["case_failed"],
        max_version=skill_results[0]["case_version"],
        max_version_case_total=skill_results[0]["case_total"],
        max_version_failed_total=skill_results[0]["case_failed"],
        run_total=skill_results[0]["case_total"] + qa_total,
        run_failed=skill_results[0]["case_failed"] + qa_failed
    )
    logging.info(smartest_result)

    instance_m8 = SmartVoiceM8()

    nlp_result = NLP_TEMPLATE.format(
        access_total=instance_m8.nlp_access_count(),
        max_qps=int(instance_m8.nlp_max_qps()),
        cost_avg=int(instance_m8.nlp_cost_avg())
    )
    logging.info(nlp_result)

    asr_ctrl_result = ASR_TEMPLATE.format(
        access_total=instance_m8.asr_access_count(),
        max_qps=int(instance_m8.asr_max_qps()),
        cost_avg=int(instance_m8.asr_cost_avg())
    )
    logging.info(asr_ctrl_result)

    feishu_url = "https://open.feishu.cn/open-apis/bot/v2/hook/75df3e29-9a07-40c2-8737-a79c658f4705"
    text = f"""
周wiki数据更新[{instance_m8.datetime_7_days_ago.strftime('%Y/%m/%d')} ~ {instance_m8.datetime_today.strftime('%Y/%m/%d')}]

{asr_ctrl_result}

{nlp_result}

{smartest_result}
""".strip()
    requests.request(method="POST",
                     url=feishu_url,
                     json={
                         "msg_type": "text",
                         "content": json.dumps({"text": text}, ensure_ascii=False)
                     })


if __name__ == '__main__':
    scheduler = BlockingScheduler()
    # mon tue wed thu fri stat sun
    scheduler.add_job(main, "cron", day_of_week="wed", hour="10", minute="0", second="0")
    logging.info("start success")
    scheduler.start()
