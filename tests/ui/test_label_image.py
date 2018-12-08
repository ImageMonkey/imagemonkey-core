from selenium import webdriver
import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver

class TestLabelImage(unittest.TestCase):
	def setUp(self):
		self._driver = ImageMonkeyChromeWebDriver()
		self._client = ImageMonkeyWebClient(self._driver)

	@classmethod
	def setUpClass(cls):
		helper.initialize_with_moderator()

	def tearDown(self):
		self._driver.quit()

	def test_label_image_should_succeed(self):
		self._client.donate("C:\\imagemonkey-core\\tests\\images\\apples\\apple2.jpeg", "apple", True)
		self._client.login("moderator", "moderator", True)
		self._client.unlock_image()
		self._client.label_image(["floor", "wall"])
		image_labels = self._client.image_labels()
		self.assertListEqual(sorted(image_labels), sorted(["apple", "wall", "floor"]))