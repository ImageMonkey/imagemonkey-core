import os
import unittest
import helper
from client import ImageMonkeyWebClient
from webdriver import ImageMonkeyChromeWebDriver


class TestUnlockMultipleImage(unittest.TestCase):
    def setUp(self):
        self._driver = ImageMonkeyChromeWebDriver()
        self._client = ImageMonkeyWebClient(self._driver)

    @classmethod
    def setUpClass(cls):
        helper.initialize_with_moderator()

    def tearDown(self):
        self._driver.quit()

    def test_unlock_image_should_succeed(self):
        path = os.path.abspath(".." + os.path.sep +
                               "images" + os.path.sep + "apples")
        ctr = 0
        for img in os.listdir(path):
            p = path + os.path.sep + img
            print("Donate image %s" % (p,))
            self._client.donate(p, True)
            ctr += 1

            if ctr == 2:
                break

        self._client.login("moderator", "moderator", True)
        self._client.unlock_multiple_images()
