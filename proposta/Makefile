# Nome do arquivo principal
FILE = main

all: pdf clean_temp

pdf:
	# Primeira passada para gerar o .aux
	pdflatex $(FILE).tex
	# Segunda passada para resolver as referências (?)
	pdflatex $(FILE).tex

# Limpa só a sujeira, mantém o PDF
clean_temp:
	rm -f $(FILE).aux $(FILE).log $(FILE).out $(FILE).toc $(FILE).nav $(FILE).snm

# Limpa TUDO, inclusive o PDF
clean: clean_temp
	rm -f $(FILE).pdf

.PHONY: all pdf clean_temp clean
