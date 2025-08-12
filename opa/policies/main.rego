package aiven.main

import data.aiven.provider.policies.conflicting

deny contains msg if {
	# iterate over every message in the 'deny' rule
	msg := conflicting.deny[_]
}
