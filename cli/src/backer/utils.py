from rich.progress import Progress, SpinnerColumn, TextColumn

class SpinnerProgress:
    def __init__(self, text_format: str) -> None:
        spinner = SpinnerColumn()
        text_clmn = TextColumn(text_format)
        self.progress_bar = Progress(spinner, text_clmn, transient=True)
    
    def __enter__(self) -> 'SpinnerProgress':
        self.progress_bar.start()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.progress_bar.stop()

    @property
    def progressbar(self) -> Progress:
        return self.progress_bar