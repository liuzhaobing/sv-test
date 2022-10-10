# -*- coding:utf-8 -*-
import psycopg2


class DataBasePostGre:
    def __init__(self, db_info):
        self.conn = psycopg2.connect(host=db_info['host'], port=db_info['port'], user=db_info['user'],
                                     password=db_info['password'], database=db_info['dbname'])
        self.cursor = self.conn.cursor()

    def query(self, query_string):
        self.cursor.execute(query_string)
        self.conn.commit()
        return self.cursor.fetchall()

    def query_without_results(self, query_string):
        self.cursor.execute(query_string)
        return self.conn.commit()

    def __del__(self):
        self.cursor.close()
        self.conn.close()
