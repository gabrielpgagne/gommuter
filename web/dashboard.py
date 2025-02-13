# Import packages
from dash import Dash, html, dcc, Input, Output
import pandas as pd
import plotly.express as px


def load_commute_time(file):
    try:
        df = pd.read_csv(file, names=["datetime", "commute_time"])
        df_dt = pd.to_datetime(df["datetime"])
        df["hour_min"] = df_dt.dt.strftime("%H:%M")
        df.sort_values("hour_min", inplace=True)
    except Exception:
        df = pd.DataFrame({"hour_min": [], "commute_time": []})
    return df


def create_histogram(df):
    return px.histogram(
        df,
        x="hour_min",
        y="commute_time",
        text_auto=True,
        histfunc="avg",
        labels={
            "hour_min": "Time of day",
            "commute_time": "Commute time (minutes)",
        },
    )


# Initialize the app
app = Dash()

# App layout
app.layout = [
    html.Div(children="Commuting time dashboard", style={"textAlign": "center"}),
    html.Hr(),
    # TODO select days of the week to display
    # dcc.Checklist(
    #     ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"],
    #     ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"],
    #     # inline=True,
    #     persistence=True,
    # ),
    # html.Hr(),
    # dash_table.DataTable(data=df_to.to_dict("records"), page_size=10),
    dcc.Interval(
        id="interval-component",
        interval=10 * 1000 * 60,  # 10 minutes
    ),
    dcc.Graph(
        figure=create_histogram(load_commute_time("data/to.csv")),
        id="hist-to",
    ),
    dcc.Graph(
        figure=create_histogram(load_commute_time("data/from.csv")),
        id="hist-from",
    ),
]


@app.callback(
    Output("hist-to", "figure"),
    Input("interval-component", "n_intervals"),
)
def update_to_work_graph(n):
    df = load_commute_time("data/to.csv")
    return create_histogram(df)


@app.callback(
    Output("hist-from", "figure"),
    Input("interval-component", "n_intervals"),
)
def update_from_work_graph(n):
    df = load_commute_time("data/from.csv")
    return create_histogram(df)


# Run the app
if __name__ == "__main__":
    app.run(host="0.0.0.0", debug=True)
