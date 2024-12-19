describe('Dashboard', () => {
    context('confirmed identity', () => {
        it('shows the certificate provider card', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=confirmYourIdentity');

            cy.contains('li', 'Confirm your details').should('contain', 'Completed');
            cy.contains('li', 'Confirm your identity').should('contain', 'Completed');
            cy.contains('li', 'Provide your certificate').should('contain', 'Not started');

            cy.visit('/dashboard')

            cy.contains('I’m a certificate provider').click()
            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.contains('a', 'Go to task list')
        })
    })

    context('provided certificate', () => {
        it('does not show the certificate provider card', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=provideYourCertificate');

            cy.contains('li', 'Confirm your details').should('contain', 'Completed');
            cy.contains('li', 'Confirm your identity').should('contain', 'Completed');
            cy.contains('li', 'Provide your certificate').should('contain', 'Completed');

            cy.visit('/dashboard')

            cy.contains('I’m a certificate provider').should('not.exist')
        })
    })

    context('provided certificate but identity mismatch', () => {
        it('shows the certificate provider card', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=provideYourCertificate&idStatus=mismatch');

            cy.contains('li', 'Confirm your details').should('contain', 'Completed');
            cy.contains('li', 'Confirm your identity').should('contain', 'Pending');
            cy.contains('li', 'Provide your certificate').should('contain', 'Completed');

            cy.visit('/dashboard')

            cy.contains('I’m a certificate provider').click()
            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.contains('a', 'Go to task list')
        })
    })
})
