from apscheduler.schedulers.blocking import BlockingScheduler
import logging

from badcase_tagging import badcase_tagging_push, badcase_tagging_pull
from cases_count import cases_construction
from qa_sync import CMSSync

logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
fh = logging.FileHandler(filename=f'crontab.log', mode="a", encoding="utf-8")
fh.setLevel(logging.INFO)
formatter = logging.Formatter('%(asctime)s [%(levelname)s] [%(module)s] [%(funcName)s]: %(message)s')
fh.setFormatter(formatter)
logger.addHandler(fh)
console = logging.StreamHandler()
console.setLevel(logging.INFO)
logger.addHandler(console)

logger.info("logger started success!")


def auto_push_bad_case():
    result, task_name = badcase_tagging_push()
    logger.info("badcase_tagging_push result: ", result)
    logger.info("badcase_tagging_push task_name: ", task_name)


def auto_pull_bad_case():
    ids, cases = badcase_tagging_pull()
    logger.info("badcase_tagging_pull ids: ", ids)
    # logger.info("badcase_tagging_pull cases: ", cases)


def auto_sync_qa_case():
    logger.info("auto sync qa case from cms start!")
    CMSSync().qa_sync(3)
    logger.info("auto sync qa case from cms end!")


if __name__ == '__main__':
    scheduler = BlockingScheduler()
    scheduler.add_job(auto_push_bad_case, "cron", day_of_week="tue", hour="9", minute="0", second="30",
                      timezone="Asia/Shanghai")
    scheduler.add_job(auto_pull_bad_case, "cron", day_of_week="tue", hour="9", minute="0", second="30",
                      timezone="Asia/Shanghai")
    scheduler.add_job(auto_sync_qa_case, "cron", hour="1", minute="0", second="30",
                      timezone="Asia/Shanghai")
    scheduler.add_job(cases_construction, "cron", hour="15", minute="0", second="0",
                      timezone="Asia/Shanghai")
    scheduler.start()
