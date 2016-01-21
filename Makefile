# This is the Makefile helping you submit the labs.  
# Just create 6.824/api.key with your API key in it, 
# and submit your lab with the following command: 
#     $ make [lab1|lab2a|lab2b|lab3a|lab3b|lab4a|lab4b|lab5]

KEY=$(shell cat api.key)
LABS=" lab1 lab2a lab2b lab3a lab3b lab4a lab4b lab5 "

%:
	@if echo $(LABS) | grep -q " $@ " ; then \
	    tar cvzf $@-handin.tar.gz --exclude=src/main/kjv12.txt Makefile .git src; \
	    if test -z $(KEY) ; then \
	        echo "Missing $(PWD)/api.key. Please create the file with your key in it or submit the $@-handin.tar.gz via the web interface."; \
	    else \
                echo "Are you sure you want to submit $@? Enter 'yes' to continue:"; \
                read line; \
                if test $$line != "yes" ; then echo "Giving up submission"; exit; fi; \
                if test `stat -c "%s" "$@-handin.tar.gz" 2>/dev/null || stat -f "%z" "$@-handin.tar.gz"` -ge 20971520 ; then echo "File exceeds 20MB."; exit; fi; \
	        curl -F file=@$@-handin.tar.gz -F key=$(KEY) http://6824.scripts.mit.edu/submit/handin.py/upload; \
	    fi; \
        else \
            echo "Bad target $@. Usage: make [$(LABS)]"; \
        fi
