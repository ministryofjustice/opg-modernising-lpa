describe('Sign', () => {
    describe('as an attorney', () => {
        beforeEach(() => {
            cy.visit('/fixtures/attorney?redirect=/sign&progress=readTheLPA');
        });

        it('can be signed', () => {
            cy.checkA11yApp();

            cy.contains('Sign as an attorney on this LPA');

            cy.contains('label', 'I, Jessie Jones, confirm').click();
            cy.contains('button', 'Submit signature').click();

            cy.url().should('contain', '/what-happens-next');
            cy.checkA11yApp();

            cy.contains('h1', 'You’ve formally agreed to be an attorney');
        });

        it('can be opted out of', () => {
            cy.contains('a', 'I do not want to be an attorney').click();

            cy.url().should('contain', '/confirm-you-do-not-want-to-be-an-attorney');
            cy.checkA11yApp();
            cy.contains('button', 'Confirm').click();

            cy.url().should('contain', '/you-have-decided-not-to-be-an-attorney');
            cy.checkA11yApp();
        });

        it('shows an error when not selected', () => {
            cy.contains('button', 'Submit signature').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('You must select the box to sign as an attorney');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'You must select the box to sign as an attorney');
        });

        it('shows an error when the wrong language is used', () => {
            cy.contains('a', 'Cymraeg').click();

            cy.contains('label', 'Rydw i, Jessie Jones, yn cadarnhau').click();
            cy.contains('button', 'Cyflwyno llofnod').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('yn Saesneg');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'yn Saesneg');
        });
    });

    describe('as a replacement attorney', () => {
        beforeEach(() => {
            cy.visit('/fixtures/attorney?redirect=/sign&options=is-replacement&progress=readTheLPA');
        });

        it('can be signed', () => {
            cy.checkA11yApp();

            cy.contains('Sign as a replacement attorney on this LPA');

            cy.contains('label', 'I, Blake Buckley, confirm').click();
            cy.contains('button', 'Submit signature').click();

            cy.url().should('contain', '/what-happens-next');
            cy.checkA11yApp();

            cy.contains('h1', 'You’ve formally agreed to be a replacement attorney');
        });

        it('can be opted out of', () => {
            cy.contains('a', 'I do not want to be an attorney').click();

            cy.url().should('contain', '/confirm-you-do-not-want-to-be-an-attorney');
            cy.checkA11yApp();
            cy.contains('button', 'Confirm').click();

            cy.url().should('contain', '/you-have-decided-not-to-be-an-attorney');
            cy.checkA11yApp();
        });

        it('shows an error when not selected', () => {
            cy.contains('button', 'Submit signature').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('You must select the box to sign as a replacement attorney');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'You must select the box to sign as a replacement attorney');
        });

        it('shows an error when the wrong language is used', () => {
            cy.contains('a', 'Cymraeg').click();

            cy.contains('label', 'Rydw i, Blake Buckley, yn cadarnhau').click();
            cy.contains('button', 'Cyflwyno llofnod').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('yn Saesneg');
            });

            cy.contains('.govuk-form-group .govuk-error-message', 'yn Saesneg');
        });
    });
});
