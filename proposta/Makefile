FILE = main

all: pdf clean_temp

pdf:
	pdflatex $(FILE).tex
	pdflatex $(FILE).tex

clean_temp:
	rm -f $(FILE).aux $(FILE).log $(FILE).out $(FILE).toc $(FILE).nav $(FILE).snm

clean: clean_temp
	rm -f $(FILE).pdf

.PHONY: all pdf clean_temp clean
