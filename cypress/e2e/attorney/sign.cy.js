describe('Sign', () => {
    describe('as an attorney', () =>{
        beforeEach(() => {
            cy.visit('/testing-start?cookiesAccepted=1&redirect=/attorney-sign&completeLpa=1&withAttorney=1&asAttorney=1');
        });

        it('can be signed', () => {
            cy.checkA11yApp();

            cy.contains('Sign as an attorney on this LPA');

            cy.contains('label', 'I, John Smith, confirm').click();
            cy.contains('button', 'Submit signature').click();

            cy.url().should('contain', '/attorney-next-page');
        });

        it('shows an error when not selected', () => {
            cy.contains('button', 'Submit signature').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select confirm');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'Select confirm');
        });
    });

    describe('as a replacement attorney', () =>{
        beforeEach(() => {
            cy.visit('/testing-start?cookiesAccepted=1&redirect=/attorney-sign&completeLpa=1&withReplacementAttorney=1&asReplacementAttorney=1');
        });

        it('can be signed', () => {
            cy.checkA11yApp();

            cy.contains('Sign as a replacement attorney on this LPA');

            cy.contains('label', 'I, Jane Smith, confirm').click();
            cy.contains('button', 'Submit signature').click();

            cy.url().should('contain', '/attorney-next-page');
        });

        it('shows an error when not selected', () => {
            cy.contains('button', 'Submit signature').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select confirm');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'Select confirm');
        });
    });
});
