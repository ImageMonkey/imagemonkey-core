from selenium import webdriver
import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver


class TestAllPages(unittest.TestCase):
    def setUp(self):
        helper.initialize_with_moderator()
        self._driver = ImageMonkeyChromeWebDriver()
        self._client = ImageMonkeyWebClient(self._driver) 

    def test_open_all_pages_once(self):
        self._client.donate(os.path.abspath(".." + os.path.sep + "images" + os.path.sep + "apples" + os.path.sep + "apple2.jpeg"), True)
        self._client.login("moderator", "moderator", True)
        self._client.unlock_image()
        self._client.label_image(["floor", "wall"])
        
        
        endpoints = ["donate", "explore", "label?mode=default", "label?mode=browse", "label?type=image&mode=default", "label?type=image&mode=browse", "annotate?mode=default", "annotate?mode=browse", "annotate?mode=default&view=unified", "annotate?mode=browse&view=unified", "verify?mode=default", "verify?mode=browse", "verify_annotation", "refine?mode=browse", "statistics", "libraries", "models", "apps", "blog", "playground", "privacy"] #graph
        
        #check endpoints twice; once logged in and once logged out
        for i in range(2):
            if i == 1:
                self._client.logout()
                print("\n\nRun again logged out\n\n")
            for endpoint in endpoints:
                print("Testing endpoint %s" %endpoint, flush=True)
            
                try:
                    self._client.navigate_to(endpoint)
                except Exception as e:
                    # no auto annotations available, so the error is normal
                    if "annotate?add_auto_annotations=true - Failed to load resource: the server responded with a status of 422 (Unprocessable Entity)":
                        continue
                    raise e
