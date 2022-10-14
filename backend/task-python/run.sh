#!/bin/bash
nohup gunicorn -c gun.py app:app &
nohup python badcase_crontab_task.py &