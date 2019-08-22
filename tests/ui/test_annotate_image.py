import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver


class TestAnnotateImage(unittest.TestCase):
	def setUp(self):
		self._driver = ImageMonkeyChromeWebDriver()
		self._client = ImageMonkeyWebClient(self._driver)

	@classmethod
	def setUpClass(cls):
		helper.initialize_with_moderator()

	def tearDown(self):
		self._driver.quit()

	def test_annotate_image_should_succeed(self):
		path = os.path.abspath(".." + os.path.sep + "images" + os.path.sep + "apples" + os.path.sep + "apple1.jpeg")
		self._client.donate(path, True)

		self._client.login("moderator", "moderator", True)
		self._client.unlock_multiple_images()

		self._client.browse_annotate()
