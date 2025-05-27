describe('Attorney dashboard', () => {
    context('original attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'I’m an attorney').click()
            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('Jessie Jones')
        })
    })

    context('replacement attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide&options=is-replacement')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'I’m an attorney').click()
            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('Blake Buckley')
        })
    })

    context('trust corporation attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide&options=is-trust-corporation')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'I’m an attorney').click()
            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('First Choice Trust Corporation Ltd.')
        })
    })

    context('replacement trust corporation attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide&options=is-trust-corporation&options=is-replacement')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'I’m an attorney').click()
            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('Second Choice Trust Corporation Ltd.')
        })
    })
})
