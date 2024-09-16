describe('what you can do now', () => {
    context('donor failed ID check', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/what-is-vouching&progress=confirmYourIdentity&idStatus=donor:insufficient-evidence')
            cy.url().should('contain', '/what-is-vouching')
            cy.checkA11yApp()

            cy.get('input[name="yes-no"]').check('no', { force: true });
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/what-you-can-do-now')
        })

        it('can choose to get ID documents', () => {
            cy.contains('label', 'I will return to GOV.UK One Login and confirm my identity').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/task-list')
        })

        it('can choose to add a voucher', () => {
            cy.contains('label', 'I have someone who can vouch for me').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/enter-voucher')
        })

        it('can choose to withdraw LPA', () => {
            cy.contains('label', 'I no longer want to make this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/withdraw-this-lpa')
        })

        it('can choose to apply to court of protection', () => {
            cy.contains('label', 'I will apply to the Court of Protection to register this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
            cy.checkA11yApp()

            cy.contains('a', 'Return to task list').click();

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp()

            cy.contains('li', "Confirm your identity and sign the LPA").should('contain', 'In progress').click();

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
            cy.checkA11yApp()

            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/read-your-lpa')
            cy.checkA11yApp()
        })

        it('errors when option not selected', () => {
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/what-you-can-do-now')
            cy.checkA11yApp()

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select what you would like to do');
            });

            cy.contains('.govuk-error-message', 'Select what you would like to do');
        })
    })

    context('voucher failed ID check', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now&progress=confirmYourIdentity&idStatus=voucher:insufficient-evidence&failedVouchAttempts=1')
            cy.url().should('contain', '/what-you-can-do-now')
        })

        it('provides next steps', () => {
            cy.contains('h2', 'Try vouching again')
            cy.contains('label', 'I have someone else who can vouch for me').click()
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/enter-voucher')
        })
    })

    context('two failed vouch attempts', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now&progress=confirmYourIdentity&idStatus=voucher:insufficient-evidence&failedVouchAttempts=2')
            cy.url().should('contain', '/what-you-can-do-now')
        })

        it('provides next steps', () => {
            cy.get('Try vouching again').should('not.exist')
            cy.get('label').should('not.contain', 'I have someone who can vouch for me')
            cy.get('label').should('not.contain', 'I have someone else who can vouch for me')

            cy.contains('label', 'I will apply to the Court of Protection to register this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')

        })
    })

})
