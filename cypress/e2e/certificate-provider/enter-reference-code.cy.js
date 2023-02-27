describe('Enter reference code', () => {
    it('can enter a valid reference code', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-reference-code').type('abcdef123456');
        cy.contains('Continue').click();

        cy.url().should('contain', '/certificate-provider-login-callback');
    });

    it('can enter a valid reference code', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-reference-code').type('abcdef123456');
        cy.contains('Continue').click();

        if (Cypress.config().baseUrl.includes('localhost')) {
            cy.url().should('contain', '/certificate-provider-login-callback');
        } else {
            cy.origin('account.gov.uk', () => {
                cy.url().should('contain', '/prove-identity-welcome');
            })
        }
    });

    it('errors when empty code', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your 12 character certificate provider reference');
        });

        cy.contains('[for=f-reference-code] ~ .govuk-error-message', 'Enter your 12 character certificate provider reference');
    });

    it('errors when incorrect code', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-reference-code').type('notATestCode');
        cy.contains('Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The certificate provider reference you entered is incorrect, please check it and try again');
        });

        cy.contains('[for=f-reference-code] ~ .govuk-error-message', 'The certificate provider reference you entered is incorrect, please check it and try again');
    });

    it('errors when incorrect code length', () => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowDonorHasPaid=1&useTestShareCode=1');

        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-reference-code').type('tooShort');
        cy.contains('Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The certificate provider reference number you enter must contain 12 characters');
        });

        cy.contains('[for=f-reference-code] ~ .govuk-error-message', 'The certificate provider reference number you enter must contain 12 characters');
    });
});
