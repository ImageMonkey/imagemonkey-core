import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver


class TestImageCollection(unittest.TestCase):
    def setUp(self):
        self._driver = ImageMonkeyChromeWebDriver()
        self._client = ImageMonkeyWebClient(self._driver)

    @classmethod
    def setUpClass(cls):
        helper.initialize_with_moderator()

    def tearDown(self):
        self._driver.quit()

    def test_create_image_collection_should_succeed(self):
        self._client.login("moderator", "moderator", True)
        self._client.create_image_collection("mycollection")
