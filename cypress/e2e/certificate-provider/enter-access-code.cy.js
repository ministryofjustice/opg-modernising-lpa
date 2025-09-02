const { TestEmail, randomAccessCode } = require("../../support/e2e");

describe('Enter access code', () => {
    context('online certificate provider', () => {
        let accessCode = ''
        beforeEach(() => {
            accessCode = randomAccessCode()

            cy.visit(`/fixtures/certificate-provider?redirect=/certificate-provider-start&withAccessCode=${accessCode}&email=${TestEmail}`);

            cy.contains('a', 'Start').click()
            cy.origin('http://localhost:7012', () => {
                cy.contains('button', 'Continue').click();
            });
            cy.url().should('contain', '/certificate-provider-enter-access-code')
        });

        it('can enter a valid access code', { pageLoadTimeout: 6000 }, () => {
            cy.checkA11yApp();

            cy.get('#f-donor-last-name').type('Smith');
            cy.get('#f-access-code').invoke('val', accessCode);
            cy.contains('Save and continue').click();

            cy.url().should('contain', '/certificate-provider-who-is-eligible')
        });

        it('errors when empty number', () => {
            cy.contains('Save and continue').click();

            cy.checkA11yApp();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter your access code');
            });

            cy.contains('[for=f-access-code] ~ .govuk-error-message', 'Enter your access code');
        });

        it('errors when incorrect code', () => {
            cy.get('#f-donor-last-name').type('Smith');
            cy.get('#f-access-code').invoke('val', 'wrongish');
            cy.contains('Save and continue').click();

            cy.checkA11yApp();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('The access code you entered is incorrect, please check it and try again');
            });

            cy.contains('[for=f-access-code] ~ .govuk-error-message', 'The access code you entered is incorrect, please check it and try again');
        });

        it('errors when incorrect code length', () => {
            cy.get('#f-donor-last-name').type('Smith');
            cy.get('#f-access-code').invoke('val', 'short');
            cy.contains('Save and continue').click();

            cy.checkA11yApp();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('The access code you enter must be 8 characters');
            });

            cy.contains('[for=f-access-code] ~ .govuk-error-message', 'The access code you enter must be 8 characters');
        });
    })

    context('paper certificate provider', () => {
        it('cannot add LPA when already submitted', () => {
            const accessCode = randomAccessCode()

            cy.visit(`/fixtures/certificate-provider?options=is-paper-donor&redirect=/certificate-provider-start&certificateProviderChannel=paper&withAccessCode=${accessCode}&email=${TestEmail}`);

            cy.contains('a', 'Start').click()
            cy.origin('http://localhost:7012', () => {
                cy.contains('label', 'Random').click();
                cy.contains('button', 'Continue').click();
            });
            cy.url().should('contain', '/certificate-provider-enter-access-code')

            cy.checkA11yApp();

            cy.get('#f-donor-last-name').type('Smith');
            cy.get('#f-access-code').invoke('val', accessCode);
            cy.contains('Save and continue').click();

            cy.url().should('contain', '/you-have-already-provided-a-certificate')
        })
    })
});
