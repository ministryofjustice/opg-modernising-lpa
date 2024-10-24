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

        it.only('can choose to apply to court of protection', () => {
            cy.contains('label', 'I will apply to the Court of Protection to register this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
            cy.checkA11yApp()

            cy.contains('a', 'Return to task list').click();

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp()

            cy.contains('li', "Confirm your identity").should('contain', 'Completed').click();

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
            cy.contains('button', 'Continue').click()

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

            cy.contains('label', 'I will apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })
    })

    context('donor ID expired', () => {
        it('provides next steps', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now-expired&progress=confirmYourIdentity&idStatus=donor:expired')
            cy.url().should('contain', '/what-you-can-do-now-expired')

            cy.contains('Your confirmed identity has expired')
            cy.contains('label', 'I will apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })
    })

    context('vouch expired', () => {
        it('provides next steps for first expired vouch', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now-expired&progress=confirmYourIdentity&idStatus=voucher:expired&failedVouchAttempts=1')
            cy.url().should('contain', '/what-you-can-do-now-expired')

            cy.contains('Your vouched-for identity has expired')
            cy.contains('h2', 'Try vouching again')

            cy.contains('label', 'I will apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })

        it('provides next steps for second expired vouch', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now-expired&progress=confirmYourIdentity&idStatus=voucher:expired&failedVouchAttempts=2')
            cy.url().should('contain', '/what-you-can-do-now-expired')

            cy.contains('Your vouched-for identity has expired');
            cy.contains('You cannot ask another person to vouch for you as only 2 attempts can be made of having someone vouch for your identity.');

            cy.get('Try vouching again').should('not.exist')
            cy.get('label').should('not.contain', 'I have someone who can vouch for me')
            cy.get('label').should('not.contain', 'I have someone else who can vouch for me')

            cy.contains('label', 'I will apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })
    })
})
