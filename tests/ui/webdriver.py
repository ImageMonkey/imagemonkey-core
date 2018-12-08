from selenium import webdriver

class ImageMonkeyChromeWebDriver(webdriver.Chrome):
	def __init__(self, headless=True, delete_all_cookies=True):
		options = webdriver.ChromeOptions()
		
		if headless:
			options.add_argument('headless')
		super(ImageMonkeyChromeWebDriver, self).__init__(chrome_options=options)

		if delete_all_cookies:
			self.delete_all_cookies()