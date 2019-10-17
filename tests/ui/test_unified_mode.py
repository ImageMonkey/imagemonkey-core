import os
import unittest
import helper
from client import ImageMonkeyWebClient
from client import UnifiedModeView
from client import RectAnnotationAction
from webdriver import ImageMonkeyChromeWebDriver


class TestUnifiedMode(unittest.TestCase):
    def setUp(self):
        helper.initialize_with_moderator()
        self._driver = ImageMonkeyChromeWebDriver()
        self._client = ImageMonkeyWebClient(self._driver) 

    def tearDown(self):
        self._driver.quit()

    def _prepare_for_test(self):
        path = os.path.abspath(".." + os.path.sep + "images" +
                               os.path.sep + "apples" + os.path.sep + "apple1.jpeg")
        self._client.donate(path, True)

        self._client.login("moderator", "moderator", True)
        self._client.unlock_multiple_images()

    def test_query_images_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["apple"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("apple", 1)

    def test_annotate_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["apple"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("apple", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-10, -15))

    def test_query_annotation_rework_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["apple"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("apple", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-10, -15))

        unified_mode_view.query("apple", 1, mode="rework")

    def test_annotate_rework_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["apple"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("apple", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-10, -15))

        unified_mode_view.query("apple", 1, mode="rework")
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-15, -15))

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("apple", 1, mode="rework")
        unified_mode_view.select_image(0) 
        unified_mode_view.check_revisions(2)

    def test_query_images_non_productive_labels_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["notexisting"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("notexisting", 1)

    def test_annotate_non_productive_label_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["notexisting"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("notexisting", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-10, -15))

    def test_query_annotation_rework_non_productive_labels_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["notexisting"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("notexisting", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-10, -15))

        unified_mode_view.query("notexisting", 1, mode="rework")

    def test_annotate_rework_non_productive_label_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["notexisting"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("notexisting", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-10, -15))

        unified_mode_view.query("notexisting", 1, mode="rework")
        unified_mode_view.select_image(0)
        unified_mode_view.annotate(RectAnnotationAction(-15, -15))

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("notexisting", 1, mode="rework")
        unified_mode_view.select_image(0) 
        unified_mode_view.check_revisions(2)

    def test_add_label_in_unified_mode_should_succeed(self):
        self._prepare_for_test()

        self._client.label_image(["apple"])

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("apple", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.label("test")

    def test_add_label_in_unified_mode_should_fail_due_to_not_authenticated(self):
        self._prepare_for_test()

        self._client.label_image(["apple"])
        self._client.logout()

        unified_mode_view = self._client.unified_mode()
        unified_mode_view.query("apple", 1)
        unified_mode_view.select_image(0)
        unified_mode_view.label("test", False)
