describe('Confirm your details', () => {
    describe('shows details for', () => {
        it('lay certificate providers', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/enter-date-of-birth');

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', '1990');

            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/your-preferred-language');
            cy.checkA11yApp();

            cy.get('[name="language-preference"]').check('cy', { force: true })

            cy.contains('button', 'Save and continue').click()

            cy.url().should('contain', '/confirm-your-details');
            cy.checkA11yApp();

            cy.contains('1 February 1990');
            cy.contains('Charlie Cooper');
            cy.contains('dt', 'Address').parent().contains('5 RICHMOND PLACE')
            cy.contains('07700 900 000');

            cy.contains('If you notice a mistake in the name, address or mobile number the donor provided');

            cy.contains('button', 'Continue').click();
            cy.url().should('contain', '/your-role');
        });

        it('professional certificate providers', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/enter-date-of-birth&options=is-professional');

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', '1990');

            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/what-is-your-home-address');
            cy.checkA11yApp();

            cy.contains('a', 'Enter address manually').click()

            cy.url().should('contain', '/what-is-your-home-address');
            cy.checkA11yApp();

            cy.get('#f-address-line-1').invoke('val', '6 RICHMOND PLACE');
            cy.get('#f-address-town').invoke('val', 'Birmingham');
            cy.get('#f-address-postcode').invoke('val', 'B14 7ED');

            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/your-preferred-language');

            cy.get('[name="language-preference"]').check('cy', { force: true })

            cy.contains('button', 'Save and continue').click()

            cy.url().should('contain', '/confirm-your-details');
            cy.checkA11yApp();

            cy.contains('1 February 1990');
            cy.contains('dt', 'Home address').parent().contains('6 RICHMOND PLACE')

            cy.contains('dt', 'Work address').parent().contains('5 RICHMOND PLACE')
            cy.contains('Charlie Cooper');
            cy.contains('5 RICHMOND PLACE');
            cy.contains('07700 900 000');

            cy.contains('button', 'Continue').click();
            cy.url().should('contain', '/your-role');
        });

        it('contact number when donor acting on paper', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/confirm-your-details&options=is-paper-donor');

            cy.contains('Contact number').should('exist');
            cy.contains('Mobile number').should('not.exist');
        })
    })

    describe('hides', () => {
        it('phone when not provided', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/confirm-your-details&options=is-paper-donor&phone=not-provided');

            cy.contains('Contact number').should('not.exist');
            cy.contains('Mobile number').should('not.exist');
            cy.contains('If you notice a mistake in the name, address or mobile number provided').should('not.exist');

            cy.contains('If you notice a mistake in the name or address provided').should('exist');
        })
    })
});
