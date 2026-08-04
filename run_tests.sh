#!/bin/sh
#
# Runs build + tests for every module in the workspace.
#
# Each workspace module is processed explicitly because "go test ./..." is
# scoped to the module where it is invoked.
#
# Linting runs in its own CI job so build and test results remain independent.

set -e

go version

# "go list -m" in workspace mode lists exactly the workspace modules and not
# their dependencies, so this stays correct as go.work changes.
modules=$(go list -m -f '{{.Dir}}')

status=0
for dir in $modules; do
	name=$(basename "$dir")
	echo "=== $name"
	if ! (
		cd "$dir"
		go build ./...
		packages=./...
		if [ "${EXCC_CI_SUPPORTED_ONLY:-}" = 1 ]; then
			packages=
			for package in $(go list ./...); do
				case "$package" in
					github.com/EXCCoin/exccd/blockchain/v4 | \
					github.com/EXCCoin/exccd/blockchain/stake/v4 | \
					github.com/EXCCoin/exccd/database/v3/ffldb | \
					github.com/EXCCoin/exccd/txscript/v4)
						continue
						;;
				esac
				packages="$packages $package"
			done
		fi
		# Intentional word splitting: packages is a go-list-generated list.
		go test -vet=all -short $packages
	); then
		echo "=== FAILED: $name"
		status=1
	fi
done

if [ "${EXCC_CI_SUPPORTED_ONLY:-}" = 1 ]; then
	go test -vet=all -short ./blockchain -run \
		'^(TestCheckProofOfWorkRejectsInvalidEquihash|TestRegularTxFirstInputValidation|TestTransactionInputValidationDeployment|TestTransactionInputValueRange|TestTransactionInputValueRangeActivation)$'
fi

if [ $status -ne 0 ]; then
	echo "------------------------------------------"
	echo "One or more modules failed."
	exit 1
fi

echo "------------------------------------------"
echo "Tests completed successfully!"
