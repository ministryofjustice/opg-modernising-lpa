describe('Sign the LPA', () => {
    describe('when certificate provider is acting online', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/task-list&progress=confirmYourIdentity');
        });

        it('can be completed ', () => {
            Cypress.on('uncaught:exception', () => {
                // TODO: remove this if this test works without, it is a problem
                // in the moj-frontend package
                return false
            })

            cy.contains('li', "Sign the LPA")
                .should('contain', 'Not started')
                .find('a')
                .click();

            cy.url().should('contain', '/how-to-sign-your-lpa');
            cy.checkA11yApp();
            cy.contains('a', 'Start').click();

            cy.url().should('contain', '/read-your-lpa');
            cy.checkA11yApp();

            cy.contains('h2', "Donor:");
            cy.contains('h2', "Attorney:");
            cy.contains('h2', "Replacement attorney:");
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/your-lpa-language');
            cy.contains('label', 'Continue and register my LPA in English').click();
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/your-legal-rights-and-responsibilities');
            cy.checkA11yApp();
            cy.contains('a', 'Continue to signing page').click();

            cy.url().should('contain', '/sign-your-lpa');
            cy.checkA11yApp();

            cy.contains('h1', "Sign your LPA");
            cy.contains('label', 'I want to sign this LPA as a deed').click();
            cy.contains('label', 'I want to apply to register this LPA').click();
            cy.contains('button', 'Submit my signature').click();

            cy.url().should('contain', '/witnessing-your-signature');
            cy.checkA11yApp();

            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/witnessing-as-certificate-provider');
            cy.checkA11yApp();

            cy.contains('h1', "Charlie Cooper, confirm you witnessed the donor sign their LPA");
            cy.get('#f-witness-code').type('1234');
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/you-have-submitted-your-lpa');
            cy.checkA11yApp();

            cy.contains('h1', "You’ve submitted your LPA");
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/dashboard');
        });

        it('errors when not signed', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=confirmYourIdentity');

            cy.visitLpa('/sign-your-lpa');

            cy.contains('button', 'Submit my signature').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select both boxes to sign and apply to register your LPA');
            });

            cy.contains('.govuk-error-message', 'Select both boxes to sign and apply to register your LPA');
        });

        it('errors when not witnessed', () => {
            cy.contains('li', "Sign the LPA")
                .should('contain', 'Not started')
                .find('a')
                .click();

            cy.contains('a', 'Start').click();
            cy.contains('a', 'Continue').click();
            cy.contains('label', 'Continue and register my LPA in English').click();
            cy.contains('button', 'Save and continue').click();
            cy.contains('a', 'Continue to signing page').click();
            cy.contains('label', 'I want to sign this LPA as a deed').click();
            cy.contains('label', 'I want to apply to register this LPA').click();
            cy.contains('button', 'Submit my signature').click();

            cy.contains('button', 'Continue').click();
            cy.contains('button', 'Continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter the code we sent to the certificate provider');
            });

            cy.contains('.govuk-error-message', 'Enter the code we sent to the certificate provider');

            cy.get('#f-witness-code').type('123');
            cy.contains('button', 'Continue').click();

            cy.contains('.govuk-error-message', 'The code we sent to the certificate provider must be 4 characters');

            cy.get('#f-witness-code').type('45');
            cy.contains('button', 'Continue').click();

            cy.contains('.govuk-error-message', 'The code we sent to the certificate provider must be 4 characters');
        });
    })

    describe('when certificate provider is acting on paper', () => {
        it('can be completed and paper forms are requested', () => {
            cy.visit('/fixtures?redirect=/read-your-lpa&progress=confirmYourIdentity&certificateProvider=paper');

            cy.url().should('contain', '/read-your-lpa');
            cy.checkA11yApp();

            cy.contains('h2', "Donor:");
            cy.contains('h2', "Attorney:");
            cy.contains('h2', "Replacement attorney:");
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/your-lpa-language');
            cy.contains('label', 'Continue and register my LPA in English').click();
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/your-legal-rights-and-responsibilities');
            cy.checkA11yApp();
            cy.contains('a', 'Continue to signing page').click();

            cy.url().should('contain', '/sign-your-lpa');
            cy.checkA11yApp();

            cy.contains('h1', "Sign your LPA");
            cy.contains('label', 'I want to sign this LPA as a deed').click();
            cy.contains('label', 'I want to apply to register this LPA').click();
            cy.contains('button', 'Submit my signature').click();

            cy.url().should('contain', '/witnessing-your-signature');
            cy.checkA11yApp();

            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/witnessing-as-certificate-provider');
            cy.checkA11yApp();

            cy.contains('h1', "Charlie Cooper, confirm you witnessed the donor sign their LPA");
            cy.get('#f-witness-code').type('1234');
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/you-have-submitted-your-lpa');
            cy.checkA11yApp();

            cy.contains('h1', "You’ve submitted your LPA");
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/dashboard');

            cy.contains('.govuk-body-s', 'Reference number:')
                .invoke('text')
                .then((text) => {
                    const uid = text.split(':')[1].trim();
                    cy.visit(`http://localhost:9001/?detail-type=paper-form-requested&detail=${uid}`);

                    cy.contains(`"uid":"${uid}"`)
                    cy.contains(`"actorType":"certificateProvider"`)
                });
        });
    })
});
