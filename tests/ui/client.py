from selenium import webdriver
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import Select
from selenium.webdriver.common.by import By
from selenium.webdriver.common.action_chains import ActionChains
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
        fn(*args, **kwargs)  # call the decorated method
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

class RectAnnotationAction(object):
    def __init__(self, x, y):
        self._x = x
        self._y = y
    
    @property
    def x(self):
        return self._x
    
    @property
    def y(self):
        return self._y

class UnifiedModeView(object):
    def __init__(self, driver):
        self._driver = driver

    def query(self, query, num_expected_images, mode="default"):
        if mode != "default" and mode != "rework":
            raise Exception("invalid mode  %s" %mode)

        elem = self._driver.find_element_by_id("annotationQuery")
        #elem = WebDriverWait(self._driver, 10).until(
        #    EC.element_to_be_clickable((By.ID, "annotationQuery"))
        #)
        
        elem.clear()
        elem.send_keys(query)

        if mode == "rework":
            self._driver.execute_script('$("#annotationsOnlyCheckbox").checkbox("set checked")');

        elem = self._driver.find_element_by_id("browseAnnotationsGoButton")
        elem.click()

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "loadingIndicator")
        wait.until(EC.invisibility_of_element_located(locator))

        time.sleep(2) #wait until images are loaded so that we can click them

        self._images = self._driver.find_elements_by_xpath(
            '//div[@id="imageGrid"]/div')
        assert len(self._images) == num_expected_images, "received expected num of images"

    def select_image(self, num):
        assert num < len(self._images), "selected image is out of range"
        assert self._images is not None, "image cannot be selected"
        self._images[num].click()


        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "loadingIndicator")
        wait.until(EC.invisibility_of_element_located(locator))

    def annotate(self, action):
        if isinstance(action, RectAnnotationAction):
            self._driver.execute_script('$("#rectMenuItem").trigger("click");')
            
            canvas = self._driver.find_element_by_id("annotationArea")
            drawing = ActionChains(self._driver)\
                .click_and_hold(canvas)\
                .move_by_offset(action.x, action.y)\
                .release()
            drawing.perform()
            self._driver.find_element_by_id("doneButton").click()
        else:
            raise Exception("error: unknown action")

    def check_revisions(self, expected_num):
        elem = self._driver.find_element_by_id("annotationRevisionsDropdownMenu") 
        children = elem.find_elements_by_tag_name("div")
        assert expected_num == len(children), "received correct num of revisions"

class ImageMonkeyWebClient(object):
    def __init__(self, driver):
        self._driver = driver
        self._username = None

    @property
    def driver(self):
        return self._driver

    @check_for_errors
    def login(self, username, password, should_succeed):
        self._driver.get(BASE_URL + "/login")
        self._driver.find_element_by_id("usernameInput").send_keys(username)
        self._driver.find_element_by_id("passwordInput").send_keys(password)
        self._driver.find_element_by_id("loginButton").click()

        if should_succeed:
            cookie = _wait_until_cookie_is_set(self._driver, 2)
            self._driver.add_cookie(cookie)
            wait = WebDriverWait(self._driver, 10)
            wait.until(EC.url_changes(BASE_URL))
            self._username = username

    @check_for_errors
    def signup(self, username, email, password):
        self._driver.get(BASE_URL + "/signup")
        self._driver.find_element_by_id("usernameInput").send_keys(username)
        self._driver.find_element_by_id("passwordInput").send_keys(password)
        self._driver.find_element_by_id(
            "repeatedPasswordInput").send_keys(password)
        self._driver.find_element_by_id("emailInput").send_keys(email)
        self._driver.find_element_by_id("signupButton").click()

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "signedUpMessage")
        wait.until(EC.visibility_of_element_located(locator))

    @check_for_errors
    def donate(self, file_path, should_succeed, image_collection=None, labels=None):
        self._driver.get(BASE_URL + "/donate")
        time.sleep(0.5)

        if image_collection is not None:
            self._driver.find_element_by_id("additionalOptionsContainer").click()
            self._driver.execute_script("$('#imageCollectionSelectionDropdown').dropdown('set selected', '%s');" %(image_collection,))

        if labels is not None:
            self._driver.execute_script("$('#labelsDropdown').dropdown('set selected', '%s');" %(','.join(labels),))

        # self._driver.execute_script("$('#labelSelector').dropdown('set selected', '%s');" %(label,))

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

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "imageGrid")
        wait.until(EC.visibility_of_element_located(locator))

        time.sleep(2) #wait until images are loaded (so that we can select them)

        images = self._driver.find_elements_by_xpath(
            '//div[@id="imageGrid"]/div')
        for image in images:
            image.click()

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "imageUnlockDoneButton")
        wait.until(EC.visibility_of_element_located(locator))

        self._driver.find_element_by_id("imageUnlockDoneButton").click()

        # after all images unlocked, open unlock page again
        # and make sure that there are no more images to unlock
        self._driver.get(BASE_URL + "/image_unlock?mode=browse")
        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "warningMessageBoxContent")
        wait.until(EC.visibility_of_element_located(locator))

        num_of_images_after_unlock = len(
            self._driver.find_elements_by_xpath('//div[@id="imageGrid"]/div'))
        assert num_of_images_after_unlock == 0, "there are still images to unlock"

    def label_image(self, labels):
        self._driver.get(BASE_URL + "/label")

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "loadingIndicator")
        wait.until(EC.invisibility_of_element_located(locator))

        # wait for intro text + dismiss it
        #locator = (By.CLASS_NAME, "shepherd-button-secondary")
        # wait.until(EC.visibility_of_element_located(locator))
        # self._driver.find_element_by_class_name("shepherd-button-secondary").click()

        # get all direct children (no grandchildren!) of div with id = 'label'
        num_of_labels_before = len(
            self._driver.find_elements_by_xpath('//div[@id="labels"]/div'))

        for label in labels:
            # Semantic UI hides the <select> and <option> tags and replaces it with nicer looking divs
            # therefore, selenium won't find the element with find_element_by_id()
            self._driver.execute_script(("addLabel('%s');" % (label,)))

            self._driver.find_element_by_xpath(
                "//button[contains(text(), 'Add')]").click()

        # get all direct children (no grandchildren!) of div with id = 'label'
        num_of_labels_after = len(
            self._driver.find_elements_by_xpath('//div[@id="labels"]/div'))
        assert (num_of_labels_before + len(labels)
                ) == num_of_labels_after, "Label Image: num of labels before do not match num of labels after"

        self._driver.find_element_by_id("doneButton").click()

    def browse_annotate(self):
        self._driver.get(BASE_URL + "/annotate?mode=browse")

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "browseAnnotationsGoButton")
        wait.until(EC.visibility_of_element_located(locator))

        elem = self._driver.find_element_by_id("annotationQuery")
        elem.clear()
        elem.send_keys("apple")

        elem = self._driver.find_element_by_id("browseAnnotationsGoButton")
        elem.click()

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "loadingIndicator")
        wait.until(EC.invisibility_of_element_located(locator))

        time.sleep(2) #wait until images are loaded so that we can click them

        images = self._driver.find_elements_by_xpath(
            '//div[@id="imageGrid"]/div')
        images[0].click()

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "loadingIndicator")
        wait.until(EC.invisibility_of_element_located(locator))

        canvas = self._driver.find_element_by_id("annotationArea")
        drawing = ActionChains(self.driver)\
            .click_and_hold(canvas)\
            .move_by_offset(-10, -15)\
            .release()
        drawing.perform()

        self._driver.find_element_by_id("doneButton").click()


    def unified_mode(self):
        self._driver.get(BASE_URL + "/annotate?mode=browse&view=unified")

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "browseAnnotationsGoButton")
        wait.until(EC.visibility_of_element_located(locator))

        return UnifiedModeView(self._driver) 

    """ 
	Returns the labels from the /label page
    """
    def image_labels(self):
        self._driver.get(BASE_URL + "/label")

        wait = WebDriverWait(self._driver, 10)
        locator = (By.ID, "loadingIndicator")
        wait.until(EC.invisibility_of_element_located(locator))

        # get all direct children (no grandchildren!) of div with id = 'label'
        labels = self._driver.find_elements_by_xpath('//div[@id="labels"]/div')
        res = []
        for label in labels:
            res.append(label.text)

        return res
    
    @check_for_errors
    def create_image_collection(self, name):
        self._driver.get(BASE_URL + "/profile/" + self._username)

        wait = WebDriverWait(self._driver, 10)
        
        locator = (By.ID, "userProfileMenuImageCollectionsTab")
        wait.until(EC.visibility_of_element_located(locator))

        table = self._driver.find_element_by_id("imageCollectionsTableContent")
        before_rows = table.find_elements(By.TAG_NAME, "tr")
        
        self._driver.find_element_by_id("userProfileMenuImageCollectionsTab").click()

        locator = (By.ID, "addImageCollectionButton")
        wait.until(EC.visibility_of_element_located(locator)) 
        
        self._driver.find_element_by_id("addImageCollectionButton").click()

        locator = (By.ID, "addImageCollectionDlg")
        wait.until(EC.visibility_of_element_located(locator)) 

        self._driver.find_element_by_id("newImageCollectionName").send_keys(name)

        self._driver.find_element_by_id("addImageCollectionDlgDoneButton").click()

        table = self._driver.find_element_by_id("imageCollectionsTableContent")
        after_rows = table.find_elements(By.TAG_NAME, "tr")
        
        failed = False
        i = 0
        while i < 5 and failed:
            failed = (len(before_rows)+1 == len(after_rows))
            time.sleep(0.5)
            i += 1

        assert not failed, "table entry not appearing"
