module github.com/f-secure-foundry/tamago/imx6/usb/ethernet

go 1.15

require (
	github.com/f-secure-foundry/tamago v0.0.0-20200819104258-eb5a3c91f51e
	golang.org/x/sys v0.0.0-20200819141100-7c7a22168250 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gvisor.dev/gvisor v0.0.0-20200819050223-35dc7fe7e78f
)

replace gvisor.dev/gvisor => github.com/f-secure-foundry/gvisor v0.0.0-20200812210008-801bb984d4b1
