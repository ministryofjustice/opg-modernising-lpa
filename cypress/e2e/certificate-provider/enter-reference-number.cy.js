describe('Enter reference number', () => {
    it('can enter a valid reference number', { pageLoadTimeout: 6000 }, () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11yApp();

        cy.get('#f-reference-number').type('abcdef123456');
        cy.contains('Continue').click();

        if (Cypress.config().baseUrl.includes('localhost')) {
            cy.url().should('contain', '/certificate-provider-check-your-name')
        } else {
            cy.origin('https://signin.integration.account.gov.uk', () => {
                cy.url().should('contain', '/')
            })
        }
    });

    it('errors when empty number', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11yApp();

        cy.contains('Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your 12 character certificate provider reference number');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'Enter your 12 character certificate provider reference number');
    });

    it('errors when incorrect code', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11yApp();

        cy.get('#f-reference-number').type('notATestCode');
        cy.contains('Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The certificate provider reference number you entered is incorrect, please check it and try again');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The certificate provider reference number you entered is incorrect, please check it and try again');
    });

    it('errors when incorrect code length', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11yApp();

        cy.get('#f-reference-number').type('tooShort');
        cy.contains('Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The certificate provider reference number you enter must contain 12 characters');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The certificate provider reference number you enter must contain 12 characters');
    });
});
