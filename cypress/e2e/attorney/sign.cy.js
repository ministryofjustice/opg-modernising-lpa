describe('Sign', () => {
    describe('as an attorney', () =>{
        beforeEach(() => {
            cy.visit('/testing-start?cookiesAccepted=1&redirect=/attorney-sign&completeLpa=1&withAttorney=1&signedByDonor=1&provideCertificate=1&asAttorney=1');
        });

        it('can be signed', () => {
            cy.checkA11yApp();

            cy.contains('Sign as an attorney on this LPA');

            cy.contains('label', 'I, John Smith, confirm').click();
            cy.contains('button', 'Submit signature').click();

            cy.url().should('contain', '/attorney-what-happens-next');
            cy.checkA11yApp();

            cy.contains('h1', 'You’ve formally agreed to be an attorney');
        });

        it('shows an error when not selected', () => {
            cy.contains('button', 'Submit signature').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select placeholder');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'Select placeholder');
        });
    });

    describe('as a replacement attorney', () =>{
        beforeEach(() => {
            cy.visit('/testing-start?cookiesAccepted=1&redirect=/attorney-sign&completeLpa=1&withReplacementAttorney=&1signedByDonor=1&provideCertificate=1&asReplacementAttorney=1');
        });

        it('can be signed', () => {
            cy.checkA11yApp();

            cy.contains('Sign as a replacement attorney on this LPA');

            cy.contains('label', 'I, Jane Smith, confirm').click();
            cy.contains('button', 'Submit signature').click();

            cy.url().should('contain', '/attorney-what-happens-next');
            cy.checkA11yApp();

            cy.contains('h1', 'You’ve formally agreed to be a replacement attorney');
        });

        it('shows an error when not selected', () => {
            cy.contains('button', 'Submit signature').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select placeholder');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'Select placeholder');
        });
    });
});
