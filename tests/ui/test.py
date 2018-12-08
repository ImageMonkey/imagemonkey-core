import os
import unittest


os.environ["PATH"] += os.pathsep + os.path.dirname(os.path.realpath(__file__))


if __name__ == '__main__':
	unittest.main(warnings='ignore')
