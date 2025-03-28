describe('You must be over 18 to complete', () => {
    it('shows your deadline when not near 18', () => {
        const dateOfBirth = new Date()
        dateOfBirth.setFullYear(dateOfBirth.getFullYear() - 17);

        cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');

        cy.contains('li', "Confirm your identity")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.url().should('contain', '/confirm-your-identity');
        cy.contains('button', 'Continue').click();

        cy.origin('http://localhost:7012', { args: { dateOfBirth } }, ({ dateOfBirth }) => {
            cy.contains('label', 'Custom').click();
            cy.get('[name=first-names]').invoke('val', 'John');
            cy.get('[name=last-name]').invoke('val', 'Johnson');
            cy.get('[name=day]').invoke('val', dateOfBirth.getDate());
            cy.get('[name=month]').invoke('val', dateOfBirth.getMonth() + 1);
            cy.get('[name=year]').invoke('val', dateOfBirth.getFullYear());
            cy.contains('button', 'Continue').click();
        });

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/you-must-be-over-18-to-complete');
        cy.checkA11yApp();

        cy.contains('You will not have turned 18 by this date');
        cy.contains('Your identity confirmation indicates you are under 18');
        cy.contains('a', 'Return to task list').click();

        cy.contains('li', "Check and send to your certificate provider")
            .should('contain', 'In progress')
            .find('a')
            .click();

        cy.contains('Your identity confirmation indicates you are under 18').should('not.exist');
    });

    it('shows your deadline when will be 18', () => {
        const dateOfBirth = new Date()
        dateOfBirth.setMonth(dateOfBirth.getMonth() - 7);
        dateOfBirth.setFullYear(dateOfBirth.getFullYear() - 17);

        cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');

        cy.contains('li', "Confirm your identity")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.url().should('contain', '/confirm-your-identity');
        cy.contains('button', 'Continue').click();

        cy.origin('http://localhost:7012', { args: { dateOfBirth } }, ({ dateOfBirth }) => {
            cy.contains('label', 'Custom').click();
            cy.get('[name=first-names]').invoke('val', 'John');
            cy.get('[name=last-name]').invoke('val', 'Johnson');
            cy.get('[name=day]').invoke('val', dateOfBirth.getDate());
            cy.get('[name=month]').invoke('val', dateOfBirth.getMonth() + 1);
            cy.get('[name=year]').invoke('val', dateOfBirth.getFullYear());
            cy.contains('button', 'Continue').click();
        });

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/you-must-be-over-18-to-complete');
        cy.checkA11yApp();

        cy.contains('You’ll turn 18 by this date');
    });
});
