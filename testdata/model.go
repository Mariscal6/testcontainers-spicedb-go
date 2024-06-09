package testdata

const MODEL = `
    definition platform {
		relation administrator: user
		permission super_admin = administrator
	}

	definition organization {
		relation platform: platform
		permission admin = platform->super_admin
	}

	definition resource {
		relation owner: user | organization
		permission admin = owner + owner->admin
	}

	definition user {}
`
