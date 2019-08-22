from selenium import webdriver
import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver


class TestDonateImage(unittest.TestCase):
	def setUp(self):
		self._driver = ImageMonkeyChromeWebDriver()
		self._client = ImageMonkeyWebClient(self._driver)

	@classmethod
	def setUpClass(cls):
		helper.clear_database()

	def test_donate_image_should_succeed(self):
		self._client.donate(os.path.abspath(".." + os.path.sep + "images" + os.path.sep + "apples" + os.path.sep + "apple2.jpeg"), True)

	def test_donate_image_should_fail(self):
		try:
			self._client.donate(os.path.abspath(".." + os.path.sep + "files" + os.path.sep + "simple-textfile.txt"), False)
		except Exception as e:
			if "the server responded with a status of 422" not in str(e):
				raise e
