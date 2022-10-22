#!/usr/bin/env python3
# -*- coding:utf-8 -*-
import json

from utils.utils_handler import Handlers
from utils.utils_mysql import DataBaseMySQL


with open("conf/config.json", "r") as f:
    config = json.load(f)


class CMSSync:
    """
    自动同步CMS的QA到测试集
    步骤：
    1.获取整个测试集
    2.获取最近7天有更改记录的线上QA
    3.遍历测试集并更新
    """

    def __init__(self):
        self.big_data = None
        self.test_cases = None

    """测试集库"""

    @staticmethod
    def qa_table(query):
        test_cases_dbinfo = config["DATABASE"]["AUTOTEST"]
        return DataBaseMySQL(test_cases_dbinfo).query(query)

    """大数据备库"""

    @staticmethod
    def big_data_table(query):
        big_data_dbinfo = config["DATABASE"]["CMS"]
        return DataBaseMySQL(big_data_dbinfo).query(query)

    """获取整个测试集"""

    def get_qa_test_cases(self):
        return self.qa_table("select id, question, answer_list from qa_base_test;")

    """获取近n天有改动的大数据"""

    def get_qa_big_data(self, n):
        n_days_ago, _ = Handlers.time_strf_n_days_ago(n)
        query = f'select id, question, answer from fqaitem ' \
                f'where is_del="no" and need_push="no" and update_time>"{n_days_ago}";'
        return self.big_data_table(query)

    """更新测试集"""

    def update_test_case(self, test_cases, big_datas):
        for test_case in test_cases:  # 遍历测试集
            for big_data in big_datas:  # 遍历大数据库
                big_data_q = big_data["question"]
                big_data_qs = json.loads(big_data_q)

                if "&&" in test_case["question"]:
                    case = test_case["question"].split("&&")[0]
                else:
                    case = test_case["question"]

                if case in big_data_qs:
                    big_data_a = big_data["answer"]
                    big_data_as = json.loads(big_data_a)
                    new_answer = "&&".join(big_data_as)
                    query = f'update qa_base_test set answer_list="{new_answer}",qa_group_id={big_data["id"]} ' \
                            f'where question="{test_case["question"]}";'
                    print(query)
                    try:
                        self.qa_table(query)
                    except Exception as e:
                        print(e)

    def qa_sync(self, n):
        """
        步骤：
        1.获取整个测试集
        2.获取最近n天有更改记录的线上QA
        3.遍历测试集并更新
        """

        all_test_cases = self.get_qa_test_cases()  # 获取整个测试集
        all_updated_big_data = self.get_qa_big_data(n)  # 获取最近n天有更改记录的线上QA
        self.update_test_case(all_test_cases, all_updated_big_data)


if __name__ == '__main__':
    tester = CMSSync()
    tester.qa_sync(10)
