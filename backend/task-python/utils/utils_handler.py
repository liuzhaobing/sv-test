#!/usr/bin/env python3
# -*- coding:utf-8 -*-
import copy
import datetime
import os
import re
import time
import json
import uuid
import sqlalchemy
import pandas as pd


class UuidHandler:
    @staticmethod
    def uuid_str():
        return str(uuid.uuid4())


class TimeHandler:
    @staticmethod
    def time_strf_now():
        """返回当前时间格式"""
        return time.strftime("%Y-%m-%d-%H-%M-%S")

    @staticmethod
    def time_now_10s():
        """返回10位时间戳 单位 秒"""
        return time.time()

    @staticmethod
    def time_strf_n_days_ago(n):
        """返回n天前的当前时间"""
        n_days_ago = (datetime.datetime.now() -
                      datetime.timedelta(days=n))
        n_days_ago_timestamp = int(time.mktime(n_days_ago.timetuple()))
        n_days_ago_strftime = n_days_ago.strftime("%Y-%m-%d %H:%M:%S")
        return n_days_ago_strftime, n_days_ago_timestamp


class ExcelHandler:

    @staticmethod
    def read_excel_as_list_map(*args, **kwargs):
        """提取全部excel数据 存入list[map]"""
        df = pd.read_excel(*args, **kwargs)
        list_map = [dict(zip(list(df.columns), line)) for line in df.values]
        return list_map

    @staticmethod
    def write_list_map_as_excel(list_map, *args, **kwargs):
        """提取全部list[map]数据 存入excel"""
        return pd.DataFrame(list_map).to_excel(*args, **kwargs)


class PyHandler:
    @staticmethod
    def list_map_duplicate_by_another_list_map(list_map_1, list_map_2, column_name):
        """根据已存在的数据list_map_1 对list_map_2进行去重"""
        lm1_cs = Handlers.read_column_as_list(list_map_1, column_name)
        new_list_map = []
        for lm in list_map_2:
            if lm[column_name] not in lm1_cs:
                new_list_map.append(lm)
        return new_list_map

    @staticmethod
    def list_map_count_and_sort(list_map, column_name):
        """按照某一列进行统计 排序"""
        values = Handlers.read_column_as_list(list_map, column_name)
        # 统计出现次数 将统计结果收集到临时map
        counter = {}
        for k in values:
            counter[k] = values.count(k)

        # 按照统计结果 开始去重 还原数据为list[map]
        new_list_map = []
        have_values = []
        for lm in list_map:
            k = lm[column_name]
            if k not in have_values:
                have_values.append(k)

                lm["counter"] = counter[k]
                new_list_map.append(lm)

        # 按照统计结果 执行倒序排列 返回排序后的list[map]
        def by_counter(i):
            return i["counter"]

        new_list_map = sorted(new_list_map, key=by_counter, reverse=True)
        return new_list_map

    @staticmethod
    def list_map_split(list_map, key):
        """按照某一列进行切割"""
        returned_list_map = []
        for m in list_map:
            keys = m[key].split("&&")
            for k in keys:
                new = copy.deepcopy(m)
                new[key] = k
                returned_list_map.append(new)
        return returned_list_map

    @staticmethod
    def list_duplicate_removal(input_list):
        """列表数据去重"""
        output_list = []
        [output_list.append(v) for v in input_list if v not in output_list]
        return output_list

    @staticmethod
    def read_column_as_list(list_map, column_name):
        """提取数据需要的值 存入list"""
        output_list = [c[column_name] for c in list_map]
        return output_list

    @staticmethod
    def list_slice_to_n(data_list, n):
        """将列表切片，按照每n个切为一组"""
        data_list_n = []
        x, y = 0, n
        while True:
            data_list_n.append(data_list[x:y:1])
            if y < len(data_list):
                x, y = y, y + n
            else:
                break
        return data_list_n

    @staticmethod
    def is_contain_chinese(string):
        """判断字符串中是否含有  中文"""
        for ch in string:
            if u'\u4e00' <= ch <= u'\u9fff':
                return True
        return False

    @staticmethod
    def is_contain_english(string):
        """判断字符串中是否含有  英文"""
        return bool(re.search('[a-zA-Z]', string))

    @staticmethod
    def is_contain_character(string):
        """判断字符串中是否含有  符号"""
        all_en_character = '!"#$%&\'()*+,-./:;<=>?@[\\]^_`{|}~'
        for s in string:
            if s in all_en_character:
                return True
        return False


class FileHandler:
    @staticmethod
    def return_file_name(file_path):
        """返回文件名"""
        return os.path.basename(file_path)

    @staticmethod
    def return_file_dir(file_path):
        """返回文件路径名"""
        return os.path.dirname(file_path)

    @staticmethod
    def make_parent_dirs(file_path):
        """创建文件父路径"""
        return os.makedirs(os.path.dirname(os.path.abspath(file_path)), exist_ok=True)

    @staticmethod
    def load_json(input_file):
        """从本地文件中加载json"""
        with open(input_file, "r", encoding="utf-8") as r:
            obj = json.load(r)
            return obj

    @staticmethod
    def save_json(obj, output_file, ensure_ascii=False, indent=4):
        """保存json到本地文件"""
        with open(output_file, "w", encoding="utf-8") as w:
            json.dump(obj, w, ensure_ascii=ensure_ascii, indent=indent)


class DBHandler:
    @staticmethod
    def database_engine(db_info):
        """
        database_info = {
            "host": "192.168.109.128",
            "user": "root",
            "password": "123456",
            "port": "3306",
            "dbname": "test",
            "dbtype": "mysql",
            "dbengine": "pymysql"
        }
        """
        try:
            sqlalchemey_database_url = '{}+{}://{}:{}@{}:{}/{}?charset=utf8'.format(
                db_info["dbtype"], db_info["dbengine"], db_info["user"], db_info["password"],
                db_info["host"], db_info["port"], db_info["dbname"]
            )
            sql_engine = sqlalchemy.create_engine(sqlalchemey_database_url)
        except Exception as e:
            raise Exception("connect to sql failed:", e)
        return sql_engine

    @staticmethod
    def excel_to_database(sql_engine, excel_path, sheet_name, db_table_name):
        """
        将Excel中数据导入到数据库表中
        :param sql_engine: 需要导入的数据库连接
        :param excel_path: 需要导入sql的excel全路径
        :param sheet_name: excel的sheet页
        :param db_table_name: 需要导入sql的数据库表名称
        """
        try:
            df = pd.read_excel(io=excel_path, sheet_name=sheet_name)
        except Exception as e:
            raise Exception("read excel failed:", e)

        try:
            df.to_sql(db_table_name, sql_engine, index=False, if_exists="append")
        except Exception as e:
            raise Exception("write sql failed:", e)

    @staticmethod
    def dict_to_database(py_dict, db_info, table_name):
        return pd.DataFrame(py_dict).to_sql(name=table_name, if_exists="append", index=False,
                                            con=DBHandler.database_engine(db_info))


class Handlers(UuidHandler, TimeHandler, ExcelHandler, PyHandler, FileHandler, DBHandler):
    pass
