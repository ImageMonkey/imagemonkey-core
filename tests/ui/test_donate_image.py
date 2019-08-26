from selenium import webdriver
import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver


class TestDonateImage(unittest.TestCase):
    def setUp(self):
        helper.initialize_with_moderator()
        self._driver = ImageMonkeyChromeWebDriver()
        self._client = ImageMonkeyWebClient(self._driver) 

    def test_donate_image_should_succeed(self):
        self._client.donate(os.path.abspath(".." + os.path.sep + "images" +
                                            os.path.sep + "apples" + os.path.sep + "apple2.jpeg"), True)
    
    def test_donate_image_ref_image_collection_should_succeed(self):
        self._client.login("moderator", "moderator", True)
        self._client.create_image_collection("mycollection")
        self._client.donate(os.path.abspath(".." + os.path.sep + "images" +
                                            os.path.sep + "apples" + os.path.sep + "apple2.jpeg"), True, image_collection="mycollection")
    

    def test_donate_image_with_labels_should_succeed(self):
        self._client.donate(os.path.abspath(".." + os.path.sep + "images" +
                                            os.path.sep + "apples" + os.path.sep + "apple2.jpeg"), True, labels=["apple"])

    def test_donate_image_should_fail(self):
        try:
            self._client.donate(os.path.abspath(
                ".." + os.path.sep + "files" + os.path.sep + "simple-textfile.txt"), False)
        except Exception as e:
            if "the server responded with a status of 422" not in str(e):
                raise e
