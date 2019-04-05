from selenium import webdriver
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.common.by import By
import subprocess
import time

BASE_URL = "http://127.0.0.1:8080"


def clear_database():
	ret = subprocess.call("go test -run TestDatabaseEmpty", shell=True, cwd="../")
	if ret != 0:
		raise Exception("Couldn't clear database")

def initialize_with_moderator():
	ret = subprocess.call("go test -run TestDatabaseEmptyWithUserThatHasUnlockImagePermission", shell=True, cwd="../")
	if ret != 0:
		raise Exception("Couldn't clear database")


