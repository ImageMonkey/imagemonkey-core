from selenium import webdriver
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import Select
from selenium.webdriver.common.by import By
from selenium.common.exceptions import WebDriverException
import subprocess
import time

BASE_URL = "http://127.0.0.1:8080"

def check_browser_errors(driver):
	try:
		browser_logs = driver.get_log('browser')
	except (ValueError, WebDriverException) as e:
		# Some browsers does not support getting logs
		print("Could not get browser logs for driver %s due to exception: %s", driver, e)
		return []

	errors = []
	for entry in browser_logs:
		if entry['level'] == 'SEVERE':
			errors.append(entry)
	return errors

def check_for_errors(fn):
	# The wrapper method which will get called instead of the decorated method:
	def wrapper(*args, **kwargs):
		fn(*args, **kwargs) # call the decorated method
		errors = check_browser_errors(args[0].driver)
		if len(errors) > 0:
			raise Exception(errors[0]['message'])

	return wrapper  # return the wrapper method


def _wait_until_cookie_is_set(driver, max_wait):
	cookie = None
	while max_wait > 0:
		cookie = driver.get_cookie("imagemonkey")
		if cookie is None:
			time.sleep(0.5)
			max_wait -= 0.5
		else:
			break

	if cookie is None:
		raise Exception("Couldn't get ImageMonkey cookie")
	return cookie


class ImageMonkeyWebClient(object):
	def __init__(self, driver):
		self._driver = driver

	@property
	def driver(self):
		return self._driver

	@check_for_errors
	def login(self, username, password, should_succeed):
		self._driver.get(BASE_URL + "/login")
		self._driver.find_element_by_id("usernameInput").send_keys(username)
		self._driver.find_element_by_id ("passwordInput").send_keys(password)
		self._driver.find_element_by_id("loginButton").click()

		if should_succeed:
			cookie = _wait_until_cookie_is_set(self._driver, 2)
			self._driver.add_cookie(cookie)
			wait = WebDriverWait(self._driver, 10)
			wait.until(EC.url_changes(BASE_URL))

	@check_for_errors
	def signup(self, username, email, password):
		self._driver.get(BASE_URL + "/signup")
		self._driver.find_element_by_id("usernameInput").send_keys(username)
		self._driver.find_element_by_id("passwordInput").send_keys(password)
		self._driver.find_element_by_id("repeatedPasswordInput").send_keys(password)
		self._driver.find_element_by_id("emailInput").send_keys(email)
		self._driver.find_element_by_id("signupButton").click()

		wait = WebDriverWait(self._driver, 10)
		locator = (By.ID, "signedUpMessage")
		wait.until(EC.visibility_of_element_located(locator))

	@check_for_errors
	def donate(self, file_path, label, should_succeed):
		self._driver.get(BASE_URL + "/donate")
		time.sleep(0.5)

		self._driver.execute_script("$('#labelSelector').dropdown('set selected', '%s');" %(label,))

		elm = self._driver.find_element_by_xpath("//input[@type='file']")
		elm.send_keys(file_path)

		wait = WebDriverWait(self._driver, 10)
		if should_succeed:
			locator = (By.ID, "successMsg")
			wait.until(EC.visibility_of_element_located(locator))
		else:
			locator = (By.ID, "failureMsg")
			wait.until(EC.visibility_of_element_located(locator))


	def unlock_image(self):
		self._driver.get(BASE_URL + "/image_unlock")

		wait = WebDriverWait(self._driver, 10)
		locator = (By.ID, "yesButton")
		wait.until(EC.visibility_of_element_located(locator))

		self._driver.find_element_by_id("yesButton").click()

		wait.until(EC.invisibility_of_element_located(locator))

	def unlock_multiple_images(self):
		self._driver.get(BASE_URL + "/image_unlock?mode=browse")

		wait = WebDriverWait(self._driver, 10)
		locator = (By.ID, "loadingIndicator")
		wait.until(EC.invisibility_of_element_located(locator))

		images = self._driver.find_elements_by_xpath('//div[@id="imageGrid"]/div')
		for image in images:
			image.click()

		self._driver.find_element_by_id("imageUnlockDoneButton").click()


	def label_image(self, labels):
		self._driver.get(BASE_URL + "/label")

		wait = WebDriverWait(self._driver, 10)
		locator = (By.ID, "loadingIndicator")
		wait.until(EC.invisibility_of_element_located(locator))

		#wait for intro text + dismiss it
		locator = (By.CLASS_NAME, "shepherd-button-secondary")
		wait.until(EC.visibility_of_element_located(locator))
		self._driver.find_element_by_class_name("shepherd-button-secondary").click()

		#get all direct children (no grandchildren!) of div with id = 'label' 
		num_of_labels_before = len(self._driver.find_elements_by_xpath('//div[@id="labels"]/div'))

		for label in labels:
			#Semantic UI hides the <select> and <option> tags and replaces it with nicer looking divs
			#therefore, selenium won't find the element with find_element_by_id()
			self._driver.execute_script(("addLabel('%s');" %(label,)))

			self._driver.find_element_by_xpath("//button[contains(text(), 'Add')]").click()

		#get all direct children (no grandchildren!) of div with id = 'label' 
		num_of_labels_after = len(self._driver.find_elements_by_xpath('//div[@id="labels"]/div'))
		assert (num_of_labels_before + len(labels)) == num_of_labels_after, "Label Image: num of labels before do not match num of labels after"

		self._driver.find_element_by_id("doneButton").click()

	""" 
	    Returns the labels from the /label page
	"""
	def image_labels(self):
		self._driver.get(BASE_URL + "/label")

		wait = WebDriverWait(self._driver, 10)
		locator = (By.ID, "loadingIndicator")
		wait.until(EC.invisibility_of_element_located(locator))

		
		#get all direct children (no grandchildren!) of div with id = 'label' 
		labels = self._driver.find_elements_by_xpath('//div[@id="labels"]/div')
		res = []
		for label in labels:
			res.append(label.text)

		return res

