# https://github.com/inversepath/tamago
#
# Copyright (c) F-Secure Corporation
# https://foundry.f-secure.com
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

APP := example
GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm
TEXT_START := 0x80010000 # ramStart (defined in imx6/imx6ul/memory.go) + 0x10000
GOFLAGS := -ldflags "-T $(TEXT_START) -E _rt0_arm_tamago -R 0x1000"
QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
        -nographic -monitor none -serial null -serial stdio -net none \
        -semihosting -d unimp

.PHONY: clean qemu qemu-gdb

all: $(APP)

$(APP):
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/inversepath/tamago-go'; \
		exit 1; \
	fi

	$(GOENV) $(TAMAGO) build $(GOFLAGS)

qemu: $(APP)
	$(QEMU) -kernel $(APP)

qemu-gdb: $(APP)
	$(QEMU) -kernel $(APP) -S -s

clean:
	rm -f $(APP)
