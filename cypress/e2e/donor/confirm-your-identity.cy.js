describe('Confirm your identity', () => {
    describe('when certificate provider is acting online', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
        });

        it('can be completed ', () => {
            cy.visitLpa("/your-details");

            cy.contains('dt', 'First names').parent().contains('a', 'Change');
            cy.contains('dt', 'Last name').parent().contains('a', 'Change');
            cy.contains('dt', 'Date of birth').parent().contains('a', 'Change');

            cy.contains('a', 'Return to task list').click();

            cy.contains('li', "Confirm your identity")
                .should('contain', 'Not started')
                .find('a')
                .click();

            cy.url().should('contain', '/confirm-your-identity');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.contains('label', 'Sam Smith (donor)').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/identity-details');
            cy.checkA11yApp();

            cy.contains('Sam');
            cy.contains('Smith');
            cy.contains('a', 'Return to task list').click();

            cy.url().should('contain', '/task-list');

            cy.visitLpa("/your-details");

            cy.contains('dt', 'First names').parent().should('not.contain', 'Change');
            cy.contains('dt', 'Last name').parent().should('not.contain', 'Change');
            cy.contains('dt', 'Date of birth').parent().should('not.contain', 'Change');
        });
    });

    describe('when insufficient evidence to prove identity', () => {
        it('can start vouching journey', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
            cy.contains('li', "Confirm your identity")
                .should('contain', 'Not started')
                .find('a')
                .click();

            cy.url().should('contain', '/confirm-your-identity');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.contains('label', 'Unable to prove identity (X)').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/unable-to-confirm-identity');
            cy.checkA11yApp();
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/choose-someone-to-vouch-for-you');
            cy.checkA11yApp();
        })
    })

    describe('when failed identity check', () => {
        it('shows problem', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
            cy.contains('li', "Confirm your identity")
                .should('contain', 'Not started')
                .find('a')
                .click();

            cy.url().should('contain', '/confirm-your-identity');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.contains('label', 'Failed identity check (T)').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/register-with-court-of-protection');
            cy.checkA11yApp();
            cy.contains('register the LPA through the Court of Protection');

            cy.contains('a', 'Return to task list').click();
            cy.contains('li', "Confirm your identity")
                .should('contain', 'There is a problem')
                .find('a')
                .click();

            cy.url().should('contain', '/register-with-court-of-protection');
        })

        it('can delete LPA', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
            cy.contains('li', "Confirm your identity")
                .should('contain', 'Not started')
                .find('a')
                .click();

            cy.url().should('contain', '/confirm-your-identity');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.contains('label', 'Failed identity check (T)').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/register-with-court-of-protection');
            cy.checkA11yApp();
            cy.contains('register the LPA through the Court of Protection');

            cy.contains('label', 'I no longer want to make this LPA').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/delete-this-lpa');
        })
    })

    describe('when identity details do not match LPA', () => {
        describe('before signing', () => {
            it('can update LPA details', () => {
                cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
                cy.contains('li', "Confirm your identity")
                    .should('contain', 'Not started')
                    .find('a')
                    .click();

                cy.url().should('contain', '/confirm-your-identity');
                cy.checkA11yApp();
                cy.contains('button', 'Continue').click();

                cy.contains('label', 'Charlie Cooper (certificate provider)').click();
                cy.contains('button', 'Continue').click();

                cy.url().should('contain', '/identity-details');
                cy.checkA11yApp();

                cy.contains('dd', 'Sam').parent().contains('span', 'Does not match');
                cy.contains('dd', 'Smith').parent().contains('span', 'Does not match');
                cy.contains('dd', '2 January 2000').parent().contains('span', 'Does not match');

                cy.contains('label', 'Yes').click();
                cy.contains('button', 'Continue').click();

                cy.url().should('contain', '/identity-details');
                cy.checkA11yApp();

                cy.contains('Your LPA details have been updated to match your confirmed identity')
                cy.get('main').should('not.contain', 'Sam');
                cy.get('main').should('not.contain', 'Smith');
                cy.get('main').should('not.contain', '2 January 2000');
            })


            it('can withdraw LPA', () => {
                cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
                cy.contains('li', "Confirm your identity")
                    .should('contain', 'Not started')
                    .find('a')
                    .click();

                cy.url().should('contain', '/confirm-your-identity');
                cy.checkA11yApp();
                cy.contains('button', 'Continue').click();

                cy.contains('label', 'Charlie Cooper (certificate provider)').click();
                cy.contains('button', 'Continue').click();

                cy.url().should('contain', '/identity-details');
                cy.checkA11yApp();

                cy.contains('dd', 'Sam').parent().contains('span', 'Does not match');
                cy.contains('dd', 'Smith').parent().contains('span', 'Does not match');
                cy.contains('dd', '2 January 2000').parent().contains('span', 'Does not match');

                cy.contains('label', 'No').click();
                cy.contains('button', 'Continue').click();

                cy.url().should('contain', '/withdraw-this-lpa');
                cy.checkA11yApp();
            })

            it('errors when option not selected', () => {
                cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
                cy.contains('li', "Confirm your identity")
                    .should('contain', 'Not started')
                    .find('a')
                    .click();

                cy.url().should('contain', '/confirm-your-identity');
                cy.checkA11yApp();
                cy.contains('button', 'Continue').click();

                cy.contains('label', 'Charlie Cooper (certificate provider)').click();
                cy.contains('button', 'Continue').click();

                cy.url().should('contain', '/identity-details');
                cy.checkA11yApp();

                cy.contains('button', 'Continue').click();

                cy.get('.govuk-error-summary').within(() => {
                    cy.contains('Select yes if you would like to update your details');
                });

                cy.contains('.govuk-error-message', 'Select yes if you would like to update your details');
            });
        });

        describe('after signing', () => {
            it('cannot update details', () => {
                cy.visit('/fixtures?redirect=/task-list&progress=signTheLpa&idStatus=donor:post-office');
                cy.contains('li', "Sign the LPA")
                    .should('contain', 'Completed');
                cy.contains('li', "Confirm your identity")
                    .should('contain', 'Pending')
                    .find('a')
                    .click();

                cy.contains('label', 'confirm my identity another way').click();
                cy.contains('button', 'Continue').click();

                cy.contains('label', 'Charlie Cooper (certificate provider)').click();
                cy.contains('button', 'Continue').click();

                cy.url().should('contain', '/identity-details');
                cy.checkA11yApp();
                cy.contains('Does not match');
                cy.contains('cannot be updated');
                cy.contains('button', 'Continue').should('not.exist');
                cy.contains('a', 'Return to task list').click();

                cy.contains('li', "Confirm your identity")
                    .should('contain', 'Pending');
            });
        });
    });

    describe('when going to the post office', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
        });

        it('can be completed ', () => {
            cy.contains('li', "Confirm your identity")
                .should('contain', 'Not started')
                .find('a')
                .click();

            cy.url().should('contain', '/confirm-your-identity');
            cy.contains('button', 'Continue').click();

            cy.go(-2);
            cy.contains('li', "Confirm your identity")
                .should('contain', 'In progress')
                .find('a')
                .click();

            cy.url().should('contain', '/how-will-you-confirm-your-identity');
            cy.checkA11yApp();
            cy.contains('label', 'I will confirm my identity at a Post Office').click();
            cy.contains('button', 'Continue').click();

            cy.contains('li', "Confirm your identity")
                .should('contain', 'Pending')
                .find('a')
                .click();

            cy.visitLpa("/your-details");

            cy.contains('dt', 'First names').parent().contains('a', 'Change');
            cy.contains('dt', 'Last name').parent().contains('a', 'Change');
            cy.contains('dt', 'Date of birth').parent().contains('a', 'Change');
        });
    });

    describe('when has invited a voucher to confirm identity', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/task-list&progress=confirmYourIdentity&idStatus=donor:insufficient-evidence&voucher=1');
        });

        it('cannot update name or date of birth', () => {
            cy.visitLpa("/your-details");

            cy.contains('dt', 'First names').parent().should('not.contain', 'Change');
            cy.contains('dt', 'Last name').parent().should('not.contain', 'Change');
            cy.contains('dt', 'Date of birth').parent().should('not.contain', 'Change');
        })
    })

    describe('when a voucher has been unable to vouch', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/task-list&progress=confirmYourIdentity&idStatus=donor:insufficient-evidence&failedVouchAttempts=1');
        });

        it('can update name and date of birth', () => {
            cy.visitLpa("/your-details");

            cy.contains('dt', 'First names').parent().should('contain', 'Change');
            cy.contains('dt', 'Last name').parent().should('contain', 'Change');
            cy.contains('dt', 'Date of birth').parent().should('contain', 'Change');
        })
    })
});
