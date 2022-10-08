# -*- coding:utf-8 -*-
import pymysql


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
