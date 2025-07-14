describe('what you can do now', () => {
    context('donor failed ID check', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/choose-someone-to-vouch-for-you&progress=confirmYourIdentity&idStatus=donor:insufficient-evidence')
            cy.url().should('contain', '/choose-someone-to-vouch-for-you')
            cy.checkA11yApp()

            cy.get('input[name="yes-no"]').check('no', { force: true });
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/what-you-can-do-now')
        })

        it('can choose to get ID documents', () => {
            cy.contains('label', 'I will return to GOV.UK One Login and confirm my identity').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/confirm-your-identity')
        })

        it('can choose to add a voucher', () => {
            cy.contains('label', 'I have someone who can vouch for me').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/enter-voucher')
        })

        it('can choose to delete LPA', () => {
            cy.contains('label', 'I no longer want to make this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/delete-this-lpa')
        })

        it('can choose to apply to court of protection', () => {
            cy.contains('label', 'Apply to the Court of Protection to register this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
            cy.checkA11yApp()

            cy.contains('a', 'Return to task list').click();

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp()

            cy.contains('li', "Confirm your identity").should('contain', 'In progress').click();

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
            cy.checkA11yApp()

            cy.contains('a', 'Return to task list').click();
            cy.url().should('contain', '/task-list')
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
            cy.visit('/fixtures?redirect=/what-you-can-do-now&progress=confirmYourIdentity&idStatus=voucher:insufficient-evidence&vouchAttempts=1')
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
            cy.visit('/fixtures?redirect=/what-you-can-do-now&progress=confirmYourIdentity&idStatus=voucher:insufficient-evidence&vouchAttempts=2')
            cy.url().should('contain', '/what-you-can-do-now')
        })

        it('provides next steps', () => {
            cy.get('Try vouching again').should('not.exist')
            cy.get('label').should('not.contain', 'I have someone who can vouch for me')
            cy.get('label').should('not.contain', 'I have someone else who can vouch for me')

            cy.contains('label', 'Apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })
    })

    context('donor ID expired', () => {
        it('provides next steps', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now-expired&progress=confirmYourIdentity&idStatus=donor:expired')
            cy.url().should('contain', '/what-you-can-do-now-expired')

            cy.contains('Your confirmed identity has expired')
            cy.contains('label', 'Apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })
    })

    context('vouch expired', () => {
        it('provides next steps for first expired vouch', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now-expired&progress=confirmYourIdentity&idStatus=voucher:expired&vouchAttempts=1')
            cy.url().should('contain', '/what-you-can-do-now-expired')

            cy.contains('Your vouched-for identity has expired')
            cy.contains('h2', 'Try vouching again')

            cy.contains('label', 'Apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })

        it('provides next steps for second expired vouch', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now-expired&progress=confirmYourIdentity&idStatus=voucher:expired&vouchAttempts=2')
            cy.url().should('contain', '/what-you-can-do-now-expired')

            cy.contains('Your vouched-for identity has expired');
            cy.contains('You cannot ask another person to vouch for you as only 2 attempts can be made of having someone vouch for your identity.');

            cy.get('Try vouching again').should('not.exist')
            cy.get('label').should('not.contain', 'I have someone who can vouch for me')
            cy.get('label').should('not.contain', 'I have someone else who can vouch for me')

            cy.contains('label', 'Apply to the Court of Protection to register this LPA').click()
            cy.contains('button', 'Continue').click()

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        })
    })

    context('want a different voucher', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/choose-someone-to-vouch-for-you&progress=payForTheLpa');
            cy.get('input[name="yes-no"]').check('yes', { force: true });
            cy.contains('button', 'Save and continue').click();
            cy.get('#f-first-names').invoke('val', 'Shopping');
            cy.get('#f-last-name').invoke('val', 'Voucher');
            cy.get('#f-email').invoke('val', 'voucher@example.com');
            cy.contains('button', 'Save and continue').click();
            cy.contains('button', 'Continue').click();
            cy.contains('a', 'Confirm my identity another way').click();
        })

        it('keeps the voucher until choice is made', () => {
            cy.visitLpa('/enter-voucher');
            cy.get('#f-first-names').should('have.value', 'Shopping');
            cy.get('#f-last-name').should('have.value', 'Voucher');
            cy.get('#f-email').should('have.value', 'voucher@example.com');
        });

        it('can choose to get ID documents', () => {
            cy.contains('label', 'I will get or find ID documents and confirm my own identity').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/are-you-sure-you-no-longer-need-voucher');
            cy.checkA11yApp();

            cy.contains('button', 'Shopping Voucher no longer needed').click();

            cy.contains('You have chosen to find, replace or get new ID');
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/confirm-your-identity')
        });

        it('can choose to add a voucher', () => {
            cy.contains('label', 'I have someone else who can vouch for me').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/are-you-sure-you-no-longer-need-voucher');
            cy.checkA11yApp();

            cy.contains('button', 'Shopping Voucher no longer needed').click();

            cy.contains('You have chosen to ask someone else');
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/enter-voucher')
        })

        it('can choose to delete LPA', () => {
            cy.contains('label', 'I no longer want to make this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/are-you-sure-you-no-longer-need-voucher');
            cy.checkA11yApp();

            cy.contains('button', 'Shopping Voucher no longer needed').click();

            cy.contains('You have told us you no longer want to make this LPA');
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/delete-this-lpa')
        })

        it('can choose to apply to court of protection', () => {
            cy.contains('label', 'Apply to the Court of Protection to register this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/are-you-sure-you-no-longer-need-voucher');
            cy.checkA11yApp();

            cy.contains('button', 'Shopping Voucher no longer needed').click();

            cy.contains('You have chosen to have your LPA reviewed by the Court of Protection');
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/what-happens-next-registering-with-court-of-protection')
        });
    });

    context('varies banner and content based on vouch status', () => {
        it('voucher has entered access code only', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now&progress=confirmYourIdentity&idStatus=donor:voucher-entered-code')

            cy.contains('Simone Sutherland has not started the process of confirming your identity');
            cy.contains('you’ll be able to nominate 1 further person to vouch for you');
        });

        it('voucher has verified donor details', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now&progress=confirmYourIdentity&idStatus=donor:verified-not-vouched')

            cy.contains('Simone Sutherland has not completed the process of confirming your identity');
            cy.contains('you’ll have to find an alternative way to confirm your identity');
        });

        it('second voucher has verified donor details', () => {
            cy.visit('/fixtures?redirect=/what-you-can-do-now&progress=confirmYourIdentity&idStatus=donor:verified-not-vouched&vouchAttempts=2')

            cy.contains('Simone Sutherland has not completed the process of confirming your identity');
            cy.contains('We suggest you contact Simone to remind them to complete their tasks');

            cy.contains('h2', 'Confirming your identity through vouching').should('not.exist');
            cy.contains('input[value=select-new-voucher]').should('not.exist');
        });
    });
})
