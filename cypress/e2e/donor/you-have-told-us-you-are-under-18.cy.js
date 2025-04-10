describe('You have told us you are under 18', () => {
    const today = new Date()

    context('when creating the first lpa', () => {
        it('can be ignored', () => {
            cy.visit('/fixtures?redirect=/your-date-of-birth');

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', today.getFullYear() - 1);
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/you-have-told-us-you-are-under-18');
            cy.checkA11yApp();

            cy.contains('a', 'Continue').click();
            cy.url().should('contain', '/do-you-live-in-the-uk');
        });

        it('can be fixed', () => {
            cy.visit('/fixtures?redirect=/your-date-of-birth');

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', today.getFullYear() - 1);
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/you-have-told-us-you-are-under-18');

            cy.contains('a', 'Change').click();
            cy.get('#f-date-of-birth-year').invoke('val', today.getFullYear() - 19);
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/do-you-live-in-the-uk');
        });
    });

    context('when the provide your details task is complete', () => {
        it('shows the warning', () => {
            cy.visit('/fixtures?progress=payForTheLpa&redirect=/your-details');
            cy.contains("dd", "2 January 2000").parent().contains("a", "Change").click()

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', today.getFullYear() - 1);
            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/you-have-told-us-you-are-under-18');
            cy.checkA11yApp();

            cy.contains('a', 'Continue').click();
            cy.url().should('contain', '/your-details');
        });
    });

    context('when creating another lpa', () => {
        it('shows the warning', () => {
            cy.visit('/fixtures/dashboard?asDonor=1&redirect=/dashboard');
            cy.contains('button', 'Start now').click();

            cy.contains("dd", "2 January 2000").parent().contains("a", "Change").click()

            cy.get('#f-date-of-birth-year').invoke('val', today.getFullYear() - 1);
            cy.contains("button", "Continue").click();

            cy.url().should('contain', '/you-have-told-us-you-are-under-18');
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/we-have-updated-your-details');
            cy.get(".govuk-notification-banner__header").contains("Date of birth updated")

            cy.contains('a', 'Continue').click();
            cy.url().should('contain', '/make-a-new-lpa');
        });
    });
});
