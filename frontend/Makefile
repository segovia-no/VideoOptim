SHELL  := /bin/bash
NVM    := [ -s "$(HOME)/.nvm/nvm.sh" ] && source "$(HOME)/.nvm/nvm.sh" && nvm use 22.12.0

.PHONY: install build dev

install:
	$(NVM) && npm install

build:
	$(NVM) && npm run build

dev:
	$(NVM) && npm run dev
