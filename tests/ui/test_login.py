import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver


class TestLogin(unittest.TestCase):
	def setUp(self):
		self._driver = ImageMonkeyChromeWebDriver()
		self._client = ImageMonkeyWebClient(self._driver)

	@classmethod
	def setUpClass(cls):
		helper.clear_database()
		client = ImageMonkeyWebClient(ImageMonkeyChromeWebDriver())
		client.signup("user", "user@imagemonkey.io", "pwd")

	def tearDown(self):
		self._driver.quit()

	def test_login_should_fail(self):
		try:
			self._client.login("not-existing-user", "pwd", False)
		except Exception as e:
			if "the server responded with a status of 401" not in str(e):
				raise e

	def test_login_should_succeed(self):
		self._client.login("user", "pwd", True)