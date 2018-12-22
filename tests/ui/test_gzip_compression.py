import os
import unittest
import helper
import requests
import client

class TestGzipCompression(unittest.TestCase):
	def setUp(self):
		pass

	@classmethod
	def setUpClass(cls):
		helper.clear_database()

	def test_gzip_compression_html_enabled(self):
		headers = {'accept-encoding':'gzip'}
		response = requests.get(client.BASE_URL + "/", headers=headers)
		self.assertEqual(response.headers["Content-Encoding"], "gzip")

	def test_gzip_compression_html_disabled(self):
		headers = {'accept-encoding':''}
		response = requests.get(client.BASE_URL + "/", headers=headers)
		
		content_encoding = None
		try:
			content_encoding = response.headers["Content-Encoding"]
		except:
			pass

		self.assertIsNone(content_encoding)

	def test_gzip_compression_css_enabled(self):
		headers = {'accept-encoding':'gzip'}
		response = requests.get(client.BASE_URL + "/css/common.css", headers=headers)
		self.assertEqual(response.headers["Content-Encoding"], "gzip")

	def test_gzip_compression_css_disabled(self):
		headers = {'accept-encoding':''}
		response = requests.get(client.BASE_URL + "/css/common.css", headers=headers)
		
		content_encoding = None
		try:
			content_encoding = response.headers["Content-Encoding"]
		except:
			pass

		self.assertIsNone(content_encoding)

	def test_gzip_compression_js_enabled(self):
		headers = {'accept-encoding':'gzip'}
		response = requests.get(client.BASE_URL + "/js/annotate.js", headers=headers)
		self.assertEqual(response.headers["Content-Encoding"], "gzip")

	def test_gzip_compression_js_disabled(self):
		headers = {'accept-encoding':''}
		response = requests.get(client.BASE_URL + "/js/annotate.js", headers=headers)
		
		content_encoding = None
		try:
			content_encoding = response.headers["Content-Encoding"]
		except:
			pass

		self.assertIsNone(content_encoding)

	#gzip compression for images is disabled. So even, if we send an 'accept-encoding': 'gzip',
	#we should get a non-gzip compressed response.
	def test_gzip_compression_img_enabled(self):
		headers = {'accept-encoding':'gzip'}
		response = requests.get(client.BASE_URL + "/img/logo.png", headers=headers)
		
		content_encoding = None
		try:
			content_encoding = response.headers["Content-Encoding"]
		except:
			pass

		self.assertIsNone(content_encoding)

	def test_gzip_compression_img_disabled(self):
		headers = {'accept-encoding':''}
		response = requests.get(client.BASE_URL + "/img/logo.png", headers=headers)
		
		content_encoding = None
		try:
			content_encoding = response.headers["Content-Encoding"]
		except:
			pass

		self.assertIsNone(content_encoding)
		