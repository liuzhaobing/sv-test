import datetime
import os

from apscheduler.schedulers.blocking import BlockingScheduler
import logging

from badcase_tagging import badcase_tagging_push, badcase_tagging_pull


def run():
    result, task_name = badcase_tagging_push()
    logger.info("badcase_tagging_push result: ", result)
    logger.info("badcase_tagging_push task_name: ", task_name)
    ids, cases = badcase_tagging_pull()
    logger.info("badcase_tagging_pull ids: ", ids)
    logger.info("badcase_tagging_pull cases: ", cases)


logger = logging.getLogger(__name__)
logger.setLevel(level=logging.INFO)
filename = os.path.abspath(__file__) + datetime.datetime.now().strftime("%Y-%m-%d-%H-%M-%S") + ".log"
file = open(filename, "w")
file.write(filename)
file.close()
handler = logging.FileHandler(filename)
handler.setLevel(logging.INFO)
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)

console = logging.StreamHandler()
console.setLevel(logging.INFO)

logger.addHandler(handler)
logger.addHandler(console)
logger.info("Start print log")
try:
    open(filename, "rb")
except (SystemExit, KeyboardInterrupt):
    raise
except Exception as e:
    logger.error("Faild to open filename from logger.error", exc_info=True)

logger.info("Finish")

if __name__ == '__main__':
    scheduler = BlockingScheduler()
    scheduler.add_job(run, "cron", day_of_week="tue", hour="9", minute="0", second="30", timezone="Asia/Shanghai")
    scheduler.start()
