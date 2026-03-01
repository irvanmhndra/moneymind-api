#!/bin/bash

# MoneyMind Database Migration Script
# Usage: ./migrate.sh [up|down|reset] [database_url]

set -e

# Default database connection (modify as needed)
DEFAULT_DB="postgresql://postgres:password@localhost:5432/moneymind"
DB_URL="${2:-$DEFAULT_DB}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if psql is available
if ! command -v psql &> /dev/null; then
    print_error "psql command not found. Please install PostgreSQL client."
    exit 1
fi

# Function to run migration up
migrate_up() {
    print_status "Running database migrations..."
    
    # Create database if it doesn't exist
    DB_NAME=$(echo $DB_URL | sed 's/.*\///g' | sed 's/\?.*//g')
    DB_URL_WITHOUT_NAME=$(echo $DB_URL | sed "s/\/$DB_NAME.*//g")
    
    print_status "Creating database '$DB_NAME' if it doesn't exist..."
    psql "$DB_URL_WITHOUT_NAME/postgres" -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || print_warning "Database '$DB_NAME' may already exist"
    
    # Run migrations in order
    for migration in $(ls migrations/*.sql | sort); do
        if [ -f "$migration" ]; then
            print_status "Applying migration: $migration"
            psql "$DB_URL" -f "$migration"
            if [ $? -eq 0 ]; then
                print_status "✓ Migration $migration completed successfully"
            else
                print_error "✗ Migration $migration failed"
                exit 1
            fi
        fi
    done
    
    print_status "All migrations completed successfully!"
}

# Function to reset database
reset_db() {
    print_warning "This will drop and recreate the entire database!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        DB_NAME=$(echo $DB_URL | sed 's/.*\///g' | sed 's/\?.*//g')
        DB_URL_WITHOUT_NAME=$(echo $DB_URL | sed "s/\/$DB_NAME.*//g")
        
        print_status "Dropping database '$DB_NAME'..."
        psql "$DB_URL_WITHOUT_NAME/postgres" -c "DROP DATABASE IF EXISTS $DB_NAME;"
        
        print_status "Recreating database '$DB_NAME'..."
        psql "$DB_URL_WITHOUT_NAME/postgres" -c "CREATE DATABASE $DB_NAME;"
        
        migrate_up
    else
        print_status "Reset cancelled."
    fi
}

# Function to show migration status
show_status() {
    print_status "Checking migration status..."
    psql "$DB_URL" -c "SELECT version, applied_at, description FROM schema_migrations ORDER BY applied_at;" 2>/dev/null || print_warning "Migrations table not found. Run 'migrate up' first."
}

# Main script logic
case "${1:-up}" in
    up)
        migrate_up
        ;;
    reset)
        reset_db
        ;;
    status)
        show_status
        ;;
    *)
        echo "Usage: $0 [up|reset|status] [database_url]"
        echo ""
        echo "Commands:"
        echo "  up     - Run pending migrations (default)"
        echo "  reset  - Drop and recreate database with all migrations"
        echo "  status - Show applied migrations"
        echo ""
        echo "Examples:"
        echo "  $0 up"
        echo "  $0 up postgresql://user:pass@localhost:5432/moneymind"
        echo "  $0 reset"
        echo "  $0 status"
        exit 1
        ;;
esac