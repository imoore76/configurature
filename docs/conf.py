# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information
from pathlib import Path

from setuptools_scm import get_version
from setuptools_scm.git import GitWorkdir

if GitWorkdir(Path("../").resolve()).get_branch() == 'stable':
	# For stable branch, just use the latest tag string
	release = get_version(
		root="../",
		version_scheme=lambda v: str(v.tag),
		local_scheme=lambda _: "",
	)
else:
	release = get_version(root='../')

project = 'Configurature'
copyright = '2024, Google LLC'
author = 'Ian Moore'

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = []

templates_path = []
exclude_patterns = []

# sphinxcontrib-osexample
extensions = [
	'myst_parser',
	'sphinx_tabs.tabs',
	'sphinx.ext.autosectionlabel',
]

autosectionlabel_prefix_document = True


# -- Options for HTML output -------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#options-for-html-output

html_theme = 'sphinx_rtd_theme'
html_static_path = []
